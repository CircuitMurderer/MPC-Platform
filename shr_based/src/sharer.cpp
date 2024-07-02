#include "utils.hpp"
#include <iostream>


std::vector<dataType> loadDataFromCsv(std::string& fileName, std::string& basic) {
    auto loadFilePath = basic + fileName;
    std::ifstream file(loadFilePath);
    std::vector<dataType> dataRows;

    if (!file.is_open()) {
        std::cerr << "Could not open the file!" << std::endl;
        return dataRows;
    }
    
    std::string line;
    std::getline(file, line);

    while (std::getline(file, line)) {
        boost::tokenizer<boost::escaped_list_separator<char>> tok(line);
        auto it = tok.begin();

        while (std::next(it, 1) != tok.end()) it++;
        dataRows.push_back(std::stod(*it));
    }
    
    file.close();
    return dataRows;
}


void shareInputData(FPOp *fpOp, int party, std::vector<dataType>& dataVec, std::string& fileName, std::string& basic) {
    auto dataLen = dataVec.size();

    std::vector<dataType> dummyVec(dataLen, 1.);
    auto dataPtr = dataVec.data();
    auto dummyPtr = dummyVec.data();

    FPArray fpArr1, fpArr2;
    BoolArray compS, compZ;
    FixArray compM, compE;

    FPArray fpArrA, fpArrB;

    if (party == ALICE_ROLE) {
        fpArr1 = fpOp->input<dataType>(ALICE_ROLE, dataLen, dataPtr);
        fpArr2 = fpOp->input<dataType>(BOB_ROLE, dataLen, dummyPtr);

        auto saveFilePath = basic + std::string("Alice") + fileName;
        std::tie(compS, compZ, compM, compE) = fpOp->get_components(fpArr1);
        serializeShare(compS.data, compZ.data, compM.data, compE.data, dataLen, saveFilePath);
    
    } else if (party == BOB_ROLE) {
        fpArr1 = fpOp->input<dataType>(ALICE_ROLE, dataLen, dummyPtr);
        fpArr2 = fpOp->input<dataType>(BOB_ROLE, dataLen, dataPtr);

        auto saveFilePath = basic + std::string("Bob") + fileName;
        std::tie(compS, compZ, compM, compE) = fpOp->get_components(fpArr2);
        serializeShare(compS.data, compZ.data, compM.data, compE.data, dataLen, saveFilePath);

    } else {
        throw std::invalid_argument("Unknown role type");
    }
}


int main(int argc, char **argv) {
    Config conf { 0, 2, 8001, "127.0.0.1", "data10k.csv", "share.bin", "result.txt", "data/" };

    ArgMapping argMap;
    argMap.arg("ro", conf.role, "Role of party: ALICE = 1; BOB = 2");
    argMap.arg("pt", conf.port, "Port Number of ALICE");
    argMap.arg("ip", conf.addr, "IP Address of ALICE");
    argMap.arg("csv", conf.csvPth, "CSV data file name");
    argMap.arg("shr", conf.shrPth, "Share data file name");
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

    auto dataVec = loadDataFromCsv(conf.csvPth, conf.bPth);
    shareInputData(fpOp, conf.role, dataVec, conf.shrPth, conf.bPth);

    auto commEnd = iopack->get_comm();
    auto duration = sci::time_from(start);

    auto dataSize = dataVec.size();
    std::cout << "Communication Cost: " << commEnd - commStart << " bytes" << std::endl;
    std::cout << "Total Time: " << duration / (1000.0) << " ms" << std::endl;

    // delete fpMath;
    delete fpOp;
    delete otpack;
    delete iopack;
}
