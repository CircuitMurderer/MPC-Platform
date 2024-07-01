#include "utils.hpp"


std::pair<FPArray, FPArray> getArrayFromBinary(FPOp *fpOp, int party, std::string& fileName, std::string& basic) {
    std::string filePath;
    if (party == ALICE_ROLE) filePath = basic + std::string("Alice") + fileName;
    else if (party == BOB_ROLE) filePath = basic + std::string("Bob") + fileName;
    else throw std::invalid_argument("Unknown role type");
    
    auto [shareS, shareZ, shareM, shareE, dataLen] = deserializeShare(filePath);
    std::vector<dataType> dummyVec(dataLen, 1.);

    FPArray fpArr1, fpArr2;
    if (party == ALICE_ROLE) {
        fpArr1 = fpOp->input(ALICE_ROLE, dataLen, shareS.data(), shareZ.data(), shareM.data(), shareE.data());
        fpArr2 = fpOp->input(BOB_ROLE, dataLen, dummyVec.data());
    } else if (party == BOB_ROLE) {
        fpArr1 = fpOp->input(ALICE_ROLE, dataLen, dummyVec.data());
        fpArr2 = fpOp->input(BOB_ROLE, dataLen, shareS.data(), shareZ.data(), shareM.data(), shareE.data());
    } else {
        throw std::invalid_argument("Unknown role type");
    }

    return std::make_pair(fpArr1, fpArr2);
}


std::vector<dataType> doCalculate(FPOp *fpOp, int party, FPArray& fpArr1, FPArray& fpArr2, Operation op) {
    auto lBound = -1 * std::numeric_limits<float>::max();
    auto rBound = std::numeric_limits<float>::max();

    FPArray fp;
    switch (op) {
    case Operation::ADD:
        std::cout << "\033[34mADD\033[0m" << std::endl;
        fp = fpOp->add(fpArr1, fpArr2);
        break;

    case Operation::SUB:
        std::cout << "\033[34mSUB\033[0m" << std::endl;
        fp = fpOp->sub(fpArr1, fpArr2);
        break;

    case Operation::MUL:
        std::cout << "\033[34mMUL\033[0m" << std::endl;
        fp = fpOp->mul(fpArr1, fpArr2);
        break;
    
    case Operation::DIV:
        std::cout << "\033[34mDIV\033[0m" << std::endl;
        fp = fpOp->div(fpArr1, fpArr2);
        break;

    case Operation::CHEAP_ADD:
        std::cout << "\033[34mCHEAP ADD\033[0m" << std::endl;
        fp = fpOp->add(fpArr1, fpArr2, true);
        break;
    
    case Operation::CHEAP_DIV:
        std::cout << "\033[34mCHEAP DIV\033[0m" << std::endl;
        fp = fpOp->div(fpArr1, fpArr2, true);
        break;
    
    default:
        throw std::invalid_argument("Unknown operation");
    }

    auto fpPub = fpOp->output(PUBLIC_ROLE, fp);
    return fpPub.get_native_type<dataType>();
}


int main(int argc, char **argv) {
    Config conf { 0, 3, 8001, "127.0.0.1", "data10k.csv", "share.bin", "result.txt", "data/" };

    ArgMapping argMap;
    argMap.arg("ro", conf.role, "Role of party: ALICE = 1; BOB = 2");
    argMap.arg("pt", conf.port, "Port Number of ALICE");
    argMap.arg("ip", conf.addr, "IP Address of ALICE");
    argMap.arg("op", conf.opt, "FP Primitve Operation");
    argMap.arg("shr", conf.shrPth, "Share data file name");
    argMap.arg("res", conf.resPth, "Result data file name");
    argMap.arg("pth", conf.bPth, "Basic data path");
    argMap.parse(argc, argv);
    auto op = static_cast<Operation>(conf.opt);
    
    auto iopack = new sci::IOPack(conf.role, conf.port, conf.addr);
    auto otpack = new sci::OTPack(iopack, conf.role);

    auto fpOp = new FPOp(conf.role, iopack, otpack);
    // auto fpMath = new FPMath(conf.role, iopack, otpack);

    auto start = sci::clock_start();
    auto commStart = iopack->get_comm();
    auto initRounds = iopack->get_rounds();

    auto [fpArr1, fpArr2] = getArrayFromBinary(fpOp, conf.role, conf.shrPth, conf.bPth);
    auto calRes = doCalculate(fpOp, conf.role, fpArr1, fpArr2, op);
    if (conf.role == ALICE_ROLE)
        saveResultToFile(calRes, conf.resPth, conf.bPth);
    
    //std::cout.precision(6);
    for (int i = 0; i < 10; i++) 
        std::cout << calRes[i] << " ";
    std::cout << std::endl;

    auto commEnd = iopack->get_comm();
    auto duration = sci::time_from(start);

    auto dataSize = calRes.size();
    std::cout << "Comm. per operations: " << 8 * (commEnd - commStart) / dataSize << " bits" << std::endl;
    std::cout << "Number of FP ops/s:\t" << (double(dataSize) / duration) * 1e6 << std::endl;
    std::cout << "Total Time:\t" << duration / (1000.0) << " ms" << std::endl;
    std::cout << "Num_rounds: " << (iopack->get_rounds() - initRounds) << std::endl;

    // delete fpMath;
    delete fpOp;
    delete otpack;
    delete iopack;
}

