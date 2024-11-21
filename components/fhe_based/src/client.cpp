#include <cstring>
#include <iomanip>
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


using namespace lbcrypto;
typedef std::string string;
typedef std::vector<double> vectorDouble;
typedef std::vector<string> vectorString;


typedef enum OperatorS {
    ADD,
    SUB,
    MUL,
    DIV,
    EXP
} Operator;


typedef struct ClientDataS {
    char id;
    int dataLen;
    Operator chosenOp;
    vectorDouble dataA;
    vectorDouble dataB;
    vectorDouble dataRes;
} ClientData;


typedef struct ServerDataS {
    int dataLen;
    Operator chosenOp;
    vectorDouble dataA;
    vectorDouble dataB;
    vectorDouble dataRes;
} ServerData;


typedef struct FilePathS {
    string basePath;
    string pubKeyPath;
    string priKeyPath;
    string varPathA;
    string varPathB;
    string varPathRes;
    string multKeyPath;
    string contextPath;
    vectorString otherPaths;
} FilePath;


typedef struct ContextConfigS {
    int multDepth;
    int scaleModSize;
    int batchSize;
    int loadCtx;
    FilePath fPath;
} ContextConfig;


class ClientProcesser {
public:
    ClientProcesser(string confPath, string dataPath) {
        opMap = buildOpMap();
        cliData = getClientData(dataPath);
        ctxConf = parseConfigFrom(confPath);

        if (ctxConf.loadCtx) {
            auto tup = loadContextAndKeyFromFile();
            ctx = std::get<CryptoContext<DCRTPoly>>(tup);
            kp = std::get<KeyPair<DCRTPoly>>(tup);
        } else {
            auto tup = genContextWithParams();
            ctx = std::get<CryptoContext<DCRTPoly>>(tup);
            kp = std::get<KeyPair<DCRTPoly>>(tup);
            saveContextAndKeyToFile();
        }

        getInfo();
    }

    std::map<string, Operator> buildOpMap() {
        return {
            {"add", ADD},
            {"sub", SUB},
            {"mul", MUL},
            {"div", DIV},
            {"exp", EXP}
        };
    }

    ClientData getClientData(string dataPath) {
        boost::property_tree::ptree pt;
        boost::property_tree::read_json(dataPath, pt);
        
        vectorDouble vectorA, vectorB, vectorRes;
        for (auto &it : pt.get_child("A")) 
            vectorA.push_back(it.second.get_value<double>());
        for (auto &it : pt.get_child("B")) 
            vectorB.push_back(it.second.get_value<double>());
        for (auto &it : pt.get_child("res")) 
            vectorRes.push_back(it.second.get_value<double>());
        
        char identify = 'N';
        auto vectorLen = pt.get<int>("dataLen");

        if (vectorA.size() == vectorLen && vectorB.size() == vectorLen) { identify = 'S'; }
        else if (vectorA.size() == vectorLen && vectorB.empty()) { identify = 'A'; } 
        else if (vectorA.empty() && vectorB.size() == vectorLen) { identify = 'B'; } 
        else {
            std::cerr << "The length of data does not match." << std::endl;
            std::exit(-1);
        }

        auto operation = opMap[pt.get<string>("chosenOp")];
        return ClientData {
            identify,
            vectorLen,
            operation,
            vectorA,
            vectorB,
            vectorRes
        };
    }

    void getInfo() {
        vectorString enumToStr = {"ADD", "SUB", "MUL", "DIV", "EXP"};

        std::cout << "=== Client Data ===" << std::endl;
        std::cout << "Client ID: " << cliData.id << std::endl;
        std::cout << "Data Length: "<< cliData.dataLen << std::endl;
        std::cout << "Chosen Operation: " << enumToStr[cliData.chosenOp] << std::endl;

        std::cout << "Data Batch A: ";
        for (auto &it : cliData.dataA) 
            std::cout << it << " ";
        std::cout << std::endl;

        std::cout << "Data Batch B: ";
        for (auto &it : cliData.dataB) 
            std::cout << it << " ";
        std::cout << std::endl;

        std::cout << "A " << enumToStr[cliData.chosenOp] << " B Results: ";
        for (auto &it : cliData.dataRes) 
            std::cout << it << " ";
        std::cout << std::endl;
    }

    void getSerDataInfo(ServerData serData) {
        vectorString enumToStr = {"ADD", "SUB", "MUL", "DIV", "EXP"};
        std::cout << "=== To Server Data ===" << std::endl;
        std::cout << "Data Length: "<< serData.dataLen << std::endl;
        std::cout << "Chosen Operation: " << enumToStr[serData.chosenOp] << std::endl;

        std::cout << "Data Batch A: ";
        for (auto &it : serData.dataA) 
            std::cout << it << " ";
        std::cout << std::endl;

        std::cout << "Data Batch B: ";
        for (auto &it : serData.dataB) 
            std::cout << it << " ";
        std::cout << std::endl;

        std::cout << "A " << enumToStr[serData.chosenOp] << " B Results: ";
        for (auto &it : serData.dataRes) 
            std::cout << it << " ";
        std::cout << std::endl;
    }

    ContextConfig parseConfigFrom(string filePath) {
        boost::property_tree::ptree pt;
        boost::property_tree::read_json(filePath, pt);

        vectorString otherPaths;
        auto fp = pt.get_child("filePaths");
        for (auto &it : fp.get_child("otherPaths")) {
            otherPaths.push_back(it.second.get_value<string>());
        }

        return ContextConfig {
            pt.get<int>("multDepth"),
            pt.get<int>("scaleModSize"),
            pt.get<int>("batchSize"),
            pt.get<int>("loadCtx"),
            FilePath {
                fp.get<string>("basePath"),
                fp.get<string>("pubKeyPath"),
                fp.get<string>("priKeyPath"),
                fp.get<string>("varPathA"),
                fp.get<string>("varPathB"),
                fp.get<string>("varPathRes"),
                fp.get<string>("multKeyPath"),
                fp.get<string>("contextPath"),
                otherPaths
            }
        };
    }

    std::tuple<CryptoContext<DCRTPoly>, KeyPair<DCRTPoly>> loadContextAndKeyFromFile() {
        CryptoContext<DCRTPoly> context;
        PublicKey<DCRTPoly> pubkey;
        PrivateKey<DCRTPoly> prikey;

        auto contextPath = ctxConf.fPath.basePath + ctxConf.fPath.contextPath;
        if (!Serial::DeserializeFromFile(contextPath, context, SerType::BINARY)) {
            std::cerr << "Error to read the crypto context." << std::endl;
            std::exit(-1);
        }

        auto pubKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.pubKeyPath;
        if (!Serial::DeserializeFromFile(pubKeyPath, pubkey, SerType::BINARY)) {
            std::cerr << "Error to read the public key." << std::endl;
            std::exit(-1);
        }

        auto priKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.priKeyPath;
        if (!Serial::DeserializeFromFile(priKeyPath, prikey, SerType::BINARY)) {
            std::cerr << "Error to read the private key." << std::endl;
            std::exit(-1);
        }

        auto multKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.multKeyPath;
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
        return std::make_tuple(context, KeyPair<DCRTPoly>(pubkey, prikey));
    }

    std::tuple<CryptoContext<DCRTPoly>, KeyPair<DCRTPoly>> genContextWithParams() {
        CCParams<CryptoContextCKKSRNS> params;
        params.SetMultiplicativeDepth(ctxConf.multDepth);
        params.SetScalingModSize(ctxConf.scaleModSize);
        params.SetBatchSize(ctxConf.batchSize);

        CryptoContext<DCRTPoly> context = GenCryptoContext(params);
        context->Enable(PKE | KEYSWITCH | LEVELEDSHE);
        std::cout << "Crypto Context generated." << std::endl;

        KeyPair<DCRTPoly> keyPair = context->KeyGen();
        std::cout << "Key Pair generated." << std::endl;

        context->EvalMultKeyGen(keyPair.secretKey);
        std::cout << "Eval Mult Keys generated." << std::endl;
        return std::make_tuple(context, keyPair);
    }

    Plaintext vecToPlaintext(vectorDouble vec) {
        return ctx->MakeCKKSPackedPlaintext(vec);
    }

    Ciphertext<DCRTPoly> encryptPlaintext(Plaintext pText) {
        return ctx->Encrypt(kp.publicKey, pText);
    }

    Plaintext decryptCiphertext(Ciphertext<DCRTPoly> cText) {
        Plaintext retText;
        ctx->Decrypt(kp.secretKey, cText, &retText);
        return retText;
    }

    ServerData calculateTransfer() {
        vectorDouble serVecA, serVecB, serVecRes;
        Operator serverOp;

        switch (cliData.chosenOp) {
        case ADD:
        case SUB:
        case MUL:   // A +.-.* B = R => A +.-.* B = R
            serverOp = cliData.chosenOp;
            serVecA = cliData.dataA;
            serVecB = cliData.dataB;
            serVecRes = cliData.dataRes;
            break;

        case DIV:   // A / B = R => A * (1 / B) = R
            serverOp = MUL;
            serVecA = cliData.dataA;
            serVecRes = cliData.dataRes;
            for (auto &it : cliData.dataB) 
                serVecB.push_back(1. / it);
            break;

        case EXP:   // A ** B = R => B * log(A) = log(R)
            serverOp = MUL;
            serVecB = cliData.dataB;
            for (auto &it : cliData.dataA) 
                serVecA.push_back(std::log2(it));
            for (auto &it : cliData.dataRes)
                serVecRes.push_back(std::log2(it));
        };

        return ServerData {
            cliData.dataLen,
            serverOp,
            serVecA,
            serVecB,
            serVecRes
        };
    }

    void transDataToServer(ServerData serData) {
        Plaintext plainA, plainB, plainRes;
        Ciphertext<DCRTPoly> cipherA, cipherB, cipherRes;

        if (!serData.dataA.empty()) {
            plainA = vecToPlaintext(serData.dataA);
            cipherA = encryptPlaintext(plainA);

            auto cipherPathA = ctxConf.fPath.basePath + ctxConf.fPath.varPathA;
            if (!Serial::SerializeToFile(cipherPathA, cipherA, SerType::BINARY)) {
                std::cerr << " Error writing ciphertext A." << std::endl;
                std::exit(-1);
            }
        }

        if (!serData.dataB.empty()) {
            plainB = vecToPlaintext(serData.dataB);
            cipherB = encryptPlaintext(plainB);

            auto cipherPathB = ctxConf.fPath.basePath + ctxConf.fPath.varPathB;
            if (!Serial::SerializeToFile(cipherPathB, cipherB, SerType::BINARY)) {
                std::cerr << " Error writing ciphertext B." << std::endl;
                std::exit(-1);
            }
        }

        /*
        if (!serData.dataRes.empty()) {
            plainRes = vecToPlaintext(serData.dataRes);
            cipherRes = encryptPlaintext(plainRes);

            auto cipherPathRes = ctxConf.fPath.basePath + ctxConf.fPath.varPathRes;
            if (!Serial::SerializeToFile(cipherPathRes, cipherRes, SerType::BINARY)) {
                std::cerr << " Error writing ciphertext Res." << std::endl;
                std::exit(-1);
            }
        }
        */
    }

    void prepareDataForServer(string serverConfig) {
        boost::property_tree::ptree pt, fp;
        auto serData = calculateTransfer();

        getSerDataInfo(serData);
        transDataToServer(serData);

        fp.put("basePath", ctxConf.fPath.basePath);
        fp.put("toCalPathA", ctxConf.fPath.varPathA);
        fp.put("toCalPathB", ctxConf.fPath.varPathB);
        fp.put("calResPath", ctxConf.fPath.varPathRes);

        fp.put("pubKeyPath", ctxConf.fPath.pubKeyPath);
        fp.put("multKeyPath", ctxConf.fPath.multKeyPath);
        fp.put("contextPath", ctxConf.fPath.contextPath);

        pt.put("toCalOp", serData.chosenOp);
        pt.add_child("filePaths", fp);
        boost::property_tree::write_json(serverConfig, pt);
    }

    void saveContextAndKeyToFile() {
        auto contextPath = ctxConf.fPath.basePath + ctxConf.fPath.contextPath;
        if (!Serial::SerializeToFile(contextPath, ctx, SerType::BINARY)) {
            std::cerr << "Error to write the crypto context." << std::endl;
            std::exit(-1);
        }

        auto pubKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.pubKeyPath;
        if (!Serial::SerializeToFile(pubKeyPath, kp.publicKey, SerType::BINARY)) {
            std::cerr << "Error to write the public key." << std::endl;
            std::exit(-1);
        }

        auto priKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.priKeyPath;
        if (!Serial::SerializeToFile(priKeyPath, kp.secretKey, SerType::BINARY)) {
            std::cerr << "Error to write the private key." << std::endl;
            std::exit(-1);
        }

        auto multKeyPath = ctxConf.fPath.basePath + ctxConf.fPath.multKeyPath;
        std::ofstream multKeyFile(multKeyPath, std::ios::out | std::ios::binary);
        if (multKeyFile.is_open()) {
            if (!ctx->SerializeEvalMultKey(multKeyFile, SerType::BINARY)) {
                std::cerr << "Error to write the eval mult key." << std::endl;
                std::exit(-1);
            }
            multKeyFile.close();
        } else {
            std::cerr << "Error to write the eval mult key." << std::endl;
            std::exit(-1);
        }

        std::cout << "Successfully serialized context and keys." << std::endl;
    }

    void getResAndVerify() {
        Ciphertext<DCRTPoly> calRes;
        auto serData = calculateTransfer();
        auto resPath = ctxConf.fPath.basePath + ctxConf.fPath.varPathRes;

        if (!Serial::DeserializeFromFile(resPath, calRes, SerType::BINARY)) {
            std::cerr << "Error to load the calculate result." << std::endl;
            std::exit(-1);
        }

        Plaintext plainRes;
        ctx->Decrypt(kp.secretKey, calRes, &plainRes);
        plainRes->SetLength(cliData.dataLen);
        auto verRes = plainRes->GetCKKSPackedValue();
        
        std::cout << "Differences between source and verified: " << std::endl;
        for (int i = 0; i < serData.dataLen; i++) 
            std::cout << std::fixed << std::setprecision(8) << std::abs(serData.dataRes[i] - verRes[i]) << " ";
        std::cout << std::endl;
    }


    void benchFileEncrypt(string filename) {
        std::vector<Ciphertext<DCRTPoly>> ciphers;
        std::ifstream file(filename, std::ios::in);
        string tempStr;

        auto t = timeNow();
        TIC(t);
        if (file.is_open()) {
            while (std::getline(file, tempStr)) {
                std::stringstream ss(tempStr);
                vectorDouble vec;
                double temp;

                while (ss >> temp) 
                    vec.push_back(temp);
                
                ciphers.push_back(
                    encryptPlaintext(
                        vecToPlaintext(vec)
                    )
                );
            }
        }
        std::cout << "[Benchmark client stage 1 on " << filename << " time usage: " << TOC_MS(t) << "ms]" << std::endl;
        TIC(t);
        for (auto &c : ciphers) {
            decryptCiphertext(c);
        }
        std::cout << "[Benchmark client stage DEC on " << filename << " time usage: " << TOC_MS(t) << "ms]" << std::endl;

        if (!ciphers.empty()) {
            if (!Serial::SerializeToFile(filename + ".fhe", ciphers, SerType::BINARY)) {
                std::cerr << " Error writing ciphertext benchs." << std::endl;
                std::exit(-1);
            }
        }
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

        //auto t = timeNow();
        for (auto &len : lens) {
            
            benchFileEncrypt(benchPath + len + "_A.txt");
            benchFileEncrypt(benchPath + len + "_B.txt");
            benchFileEncrypt(benchPath + len + "_Add.txt");
            benchFileEncrypt(benchPath + len + "_Mul.txt");
            
        }
    }

private:
    std::map<string, Operator> opMap;
    ContextConfig ctxConf;
    ClientData cliData;

    CryptoContext<DCRTPoly> ctx;
    KeyPair<DCRTPoly> kp;
};


int main(int argc, char** argv) {
    auto cp = ClientProcesser("config.json", "data_to_client.json");

    if (argc < 2) {
        std::cerr << "Missing parameters." << std::endl;
        return -1;
    }

    auto t = timeNow(); TIC(t);

    if (std::strcmp(argv[1], "make") == 0) {
        cp.prepareDataForServer("data_to_server.json");
    } else if (std::strcmp(argv[1], "verify") == 0) {
        cp.getResAndVerify();
    } else if (std::strcmp(argv[1], "bench") == 0) {
        cp.benchmark();
    } else {
        std::cerr << "Unknown parameters." << std::endl;
        return -1;
    }

    std::cout << "[Time usage: " << TOC_MS(t) << "ms]" << std::endl;
    return 0;
}