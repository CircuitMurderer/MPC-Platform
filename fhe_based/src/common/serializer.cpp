#include <iostream>

#include "ciphertext-ser.h"
#include "cryptocontext-ser.h"
#include "key/key-ser.h"
#include "scheme/ckksrns/ckksrns-ser.h"
#include "utils/debug.h"
#include "utils/serial.h"

using namespace lbcrypto;


class Serializer {
public:
    Serializer() {

    }

    bool serializeContextTo(std::string path) {
        //if (!Serial::SerializeToFile())
        return true;
    }

    bool serializePubKeyTo(std::string path) {
        return true;
    }

    bool serializeMultKeyTo(std::string path) {
        return true;
    }

    bool serializeCipherTextTo(std::string path) {
        return true;
    }

private:
    std::string basePath;
};