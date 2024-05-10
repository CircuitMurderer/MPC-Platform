#include <iostream>
#include <string>
#include <tuple>
#include <unistd.h>

#include "openfhe.h"

// header files needed for serialization
#include "key/key-ser.h"
#include "ciphertext-ser.h"
#include "cryptocontext-ser.h"
#include "scheme/ckksrns/ckksrns-ser.h"

#include <boost/property_tree/ptree.hpp>
#include <boost/property_tree/json_parser.hpp>

#define ADD 0
#define SUB 1
#define MUL 2

using namespace lbcrypto;
typedef std::string string;
typedef std::vector<double> vectorDouble;
typedef std::vector<string> vectorString;


typedef struct FilePathS {
    string basePath;
    string toCalPathA;
    string toCalPathB;
    string calResPath;
    string pubKeyPath;
    string multKeyPath;
    string contextPath;
} FilePath;


typedef struct CalConfigS {
    int toCalOp;
    FilePath fPath;
} CalConfig;


class ServerProcesser {
public:
    ServerProcesser(string confPath) {
        calConf = loadFrom(confPath);
        auto tup = getContextAndKeys();
        ctx = std::get<CryptoContext<DCRTPoly>>(tup);
        pk = std::get<PublicKey<DCRTPoly>>(tup);
    }

    CalConfig loadFrom(string confPath) {
        boost::property_tree::ptree pt, fp;
        boost::property_tree::read_json(confPath, pt);
        fp = pt.get_child("filePaths");

        return CalConfig {
            pt.get<int>("toCalOp"),
            FilePath {
                fp.get<string>("basePath"),
                fp.get<string>("toCalPathA"),
                fp.get<string>("toCalPathB"),
                fp.get<string>("calResPath"),
                fp.get<string>("pubKeyPath"),
                fp.get<string>("multKeyPath"),
                fp.get<string>("contextPath")
            }
        };
    }

    std::tuple<CryptoContext<DCRTPoly>, PublicKey<DCRTPoly>> getContextAndKeys() {
        CryptoContext<DCRTPoly> context;
        PublicKey<DCRTPoly> pubkey;

        context->ClearEvalMultKeys();
        CryptoContextFactory<DCRTPoly>::ReleaseAllContexts();

        auto ctxPath = calConf.fPath.basePath + calConf.fPath.contextPath;
        if (!Serial::DeserializeFromFile(ctxPath, context, SerType::BINARY)) {
            std::cerr << "Error to read the crypto context." << std::endl;
            std::exit(-1);
        }

        auto pubKeyPath = calConf.fPath.basePath + calConf.fPath.pubKeyPath;
        if (!Serial::DeserializeFromFile(pubKeyPath, pubkey, SerType::BINARY)) {
            std::cerr << "Error to read the public key." << std::endl;
            std::exit(-1);
        }

        auto multKeyPath = calConf.fPath.basePath + calConf.fPath.multKeyPath;
        std::ifstream multKeyFile(multKeyPath, std::ios::in | std::ios::binary);
        if (multKeyFile.is_open()) {
            if (!context->DeserializeEvalMultKey(multKeyFile, SerType::BINARY)) {
                std::cerr << "Error to read the eval mult key." << std::endl;
                std::exit(-1);
            }
            multKeyFile.close();
        } else {
            std::cerr << "Error to read the eval mult key." << std::endl;
            std::exit(-1);
        }

        std::cout << "Successfully deserialized context and keys." << std::endl;
        return std::make_tuple(context, pubkey);
    }

    Ciphertext<DCRTPoly> readCipherTextFrom(string fpath) {
        Ciphertext<DCRTPoly> cipher;
        if (!Serial::DeserializeFromFile(fpath, cipher, SerType::BINARY)) {
            std::cerr << "Cannot read cipher text from " << fpath << std::endl;
            std::exit(1);
        }
        
        return cipher;
    }

    Ciphertext<DCRTPoly> readAndCalculate() {
        auto cipherPathA = calConf.fPath.basePath + calConf.fPath.toCalPathA;
        auto cipherPathB = calConf.fPath.basePath + calConf.fPath.toCalPathB;

        auto cipherA = readCipherTextFrom(cipherPathA);
        auto cipherB = readCipherTextFrom(cipherPathB);

        switch (calConf.toCalOp) {
        case ADD:
            return ctx->EvalAdd(cipherA, cipherB);
        case SUB:
            return ctx->EvalSub(cipherA, cipherB);
        case MUL:
            return ctx->EvalMult(cipherA, cipherB);
        default:
            std::cerr << "Unknown operation." << std::endl;
            std::exit(-1);
        }
    }

    void calculateAndSave() {
        auto calRes = readAndCalculate();
        auto resPath = calConf.fPath.basePath + calConf.fPath.calResPath;

        if (!Serial::SerializeToFile(resPath, calRes, SerType::BINARY)) {
            std::cerr << "Error to save the calculate result." << std::endl;
            std::exit(-1);
        }

        std::cout << "Calculated and saved." << std::endl;
    }

    std::vector<Ciphertext<DCRTPoly>> readBenchs(string filename) {
        std::vector<Ciphertext<DCRTPoly>> ciphers;

        if (!Serial::DeserializeFromFile(filename, ciphers, SerType::BINARY)) {
            std::cerr << "Cannot read cipher text from " << filename << std::endl;
            std::exit(1);
        }

        return ciphers;
    }

    void benchmark() {
        string benchPath = "../bench_datas/";
        vectorString lens = {
            "1000",
            //"10000",
            //"100000",
            //"1000000",
            //"10000000"
        };
        auto t = timeNow();
        std::vector<Ciphertext<DCRTPoly>> as, bs, adds, muls;
        //auto t = timeNow();
        for (auto &len : lens) {
            as = readBenchs(benchPath + len + "_A.txt.fhe");
            bs = readBenchs(benchPath + len + "_B.txt.fhe");
            adds = readBenchs(benchPath + len + "_Add.txt.fhe");
            muls = readBenchs(benchPath + len + "_Mul.txt.fhe");

            std::cout << "OK\n";
            TIC(t);
            for (int i = 0; i < as.size(); i++) {
                ctx->EvalAdd(as[i], bs[i]);
            }
            std::cout << "[Benchmark ADD time usage: " << TOC_MS(t) << "ms]" << std::endl;

            TIC(t);
            for (int i = 0; i < as.size(); i++) {
                ctx->EvalMult(as[i], bs[i]);
            }
            std::cout << "[Benchmark MUL time usage: " << TOC_MS(t) << "ms]" << std::endl;
            
        }
    }

private:
    CryptoContext<DCRTPoly> ctx;
    PublicKey<DCRTPoly> pk;
    CalConfig calConf;
};

int main(int argc, char** argv) {
    auto sp = ServerProcesser("data_to_server.json");

    auto t = timeNow(); TIC(t);

    if (argc > 1) {
        sp.benchmark();
    } else {
        sp.calculateAndSave();
    }
    
    std::cout << "[Time usage: " << TOC_MS(t) << "ms]" << std::endl;

    return 0;
}