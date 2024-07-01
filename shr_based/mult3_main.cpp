// MIT License
//
// Copyright (c) 2021 Arianne Roselina Prananto
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

#include <cmath>
#include <fstream>
#include <iostream>
#include <random>
#include <regex>

#include <fmt/format.h>
#include <boost/lexical_cast.hpp>
#include <boost/program_options.hpp>

#include "base/party.h"
#include "common/mult3.h"
#include "communication/communication_layer.h"
#include "communication/tcp_transport.h"
#include "statistics/analysis.h"
#include "utility/typedefs.h"

namespace program_options = boost::program_options;

bool CheckPartyArgumentSyntax(const std::string& party_argument);

std::pair<program_options::variables_map, std::vector<bool>> ParseProgramOptions(int ac,
                                                                                 char* av[]);

encrypto::motion::PartyPointer CreateParty(const program_options::variables_map& user_options);

int main(int ac, char* av[]) {
  try {
    auto [user_options, flag] = ParseProgramOptions(ac, av);
    // if help flag is set - print allowed command line arguments and exit
    if (flag[0]) return EXIT_SUCCESS;

    encrypto::motion::MpcProtocol protocol;
    const std::string protocol_string{user_options["protocol"].as<std::string>()};
    std::map<std::string, encrypto::motion::MpcProtocol> protocol_conversion{
        {"ArithmeticGMW", encrypto::motion::MpcProtocol::kArithmeticGmw},
        {"GMW", encrypto::motion::MpcProtocol::kBooleanGmw},
        {"BooleanGMW", encrypto::motion::MpcProtocol::kBooleanGmw},
        {"BMR", encrypto::motion::MpcProtocol::kBooleanAstra},// kBmr},
    };
    bool print_output = flag[1];
    std::uint32_t input_command_line;
    std::string input_file_path, input_file_shared_path;
    if (user_options.count("input"))
      input_command_line = user_options["input"].as<std::uint32_t>();
    else if (user_options.count("input-file"))
      input_file_path = user_options["input-file"].as<std::string>();
    else
      input_file_shared_path = user_options["input-file-shared"].as<std::string>();
    encrypto::motion::AccumulatedRunTimeStatistics accumulated_statistics;
    encrypto::motion::AccumulatedCommunicationStatistics accumulated_communication_statistics;

    encrypto::motion::PartyPointer party{CreateParty(user_options)};
    // establish communication channels with other parties

    auto protocol_iterator = protocol_conversion.find(protocol_string);
    if (protocol_iterator != protocol_conversion.end()) {
      protocol = protocol_iterator->second;
      auto statistics = EvaluateProtocol(party, protocol, input_command_line, input_file_path,
                                         input_file_shared_path, print_output);
      accumulated_statistics.Add(statistics);
    } else {
      throw std::invalid_argument("Invalid MPC protocol");
    }

    auto communication_statistics =
        party->GetBackend()->GetCommunicationLayer().GetTransportStatistics();
    accumulated_communication_statistics.Add(communication_statistics);

    std::cout << encrypto::motion::PrintStatistics(fmt::format("Mult3", protocol_string),
                                                   accumulated_statistics,
                                                   accumulated_communication_statistics);

  } catch (std::runtime_error& e) {
    std::cerr << e.what() << "\n";
    return EXIT_FAILURE;
  }
  return EXIT_SUCCESS;
}

const std::regex kPartyArgumentRegex("([012]),([^,]+),(\\d{1,5})");

bool CheckPartyArgumentSyntax(const std::string& party_argument) {
  // other party's id, host address, and port
  return std::regex_match(party_argument, kPartyArgumentRegex);
}

std::tuple<std::size_t, std::string, std::uint16_t> ParsePartyArgument(
    const std::string& party_argument) {
  std::smatch match;
  std::regex_match(party_argument, match, kPartyArgumentRegex);
  auto id = boost::lexical_cast<std::size_t>(match[1]);
  auto host = match[2];
  auto port = boost::lexical_cast<std::uint16_t>(match[3]);
  return {id, host, port};
}

// <variables map, (help flag, print_output flag)>
std::pair<program_options::variables_map, std::vector<bool>> ParseProgramOptions(int ac,
                                                                                 char* av[]) {
  using namespace std::string_view_literals;
  constexpr std::string_view kConfigFileMessage =
      "configuration file, other arguments will overwrite the parameters read from the configuration file"sv;
  bool print, help, print_output;
  program_options::options_description description("Allowed options");
  // clang-format off
  description.add_options()
      ("help,h", program_options::bool_switch(&help)->default_value(false),"produce help message")
      ("disable-logging,l","disable logging to file")
      ("print-configuration,p", program_options::bool_switch(&print)->default_value(false), "print configuration")
      ("configuration-file,f", program_options::value<std::string>(), kConfigFileMessage.data())
      ("my-id", program_options::value<std::size_t>(), "my party id")
      ("parties", program_options::value<std::vector<std::string>>()->multitoken(), "(other party id, host, port, my role), e.g., --parties 0,127.0.0.1,23000 1,127.0.0.1,23001")
      ("protocol", program_options::value<std::string>()->default_value("ArithmeticGMW"), "MPC protocol")
      ("online-after-setup", program_options::value<bool>()->default_value(true), "compute the online phase of the gate evaluations after the setup phase for all of them is completed (true/1 or false/0)")
      ("print-output", program_options::bool_switch(&print_output)->default_value(false), "print result")
      ("input", program_options::value<std::uint32_t>(), "get party's input from command line, e.g. 1")
      ("input-file", program_options::value<std::string>(),
          "get party's input from file, include path e.g. ../../src/examples/tutorial/mult3/data/mult3.0.dat")
      ("input-file-shared", program_options::value<std::string>(),
          "get party's shared input from file, include path e.g. ../../src/examples/tutorial/mult3/data/mult3shared_arit.0.dat");

  // clang-format on

  program_options::variables_map user_options;

  program_options::store(program_options::parse_command_line(ac, av, description), user_options);
  program_options::notify(user_options);

  // argument help or no arguments (at least a configuration file is expected)
  if (help) {
    std::cout << description << "\n";
    return std::make_pair<program_options::variables_map, std::vector<bool>>(
        {}, std::vector<bool>{true, print_output});
  }

  // read configuration file
  if (user_options.count("configuration-file")) {
    std::ifstream user_options_file(user_options["configuration-file"].as<std::string>().c_str());
    program_options::store(program_options::parse_config_file(user_options_file, description),
                           user_options);
    program_options::notify(user_options);
  }

  // print parsed parameters
  if (user_options.count("my-id")) {
    if (print) std::cout << "My id " << user_options["my-id"].as<std::size_t>() << std::endl;
  } else
    throw std::runtime_error("My id is not set but required");

  if (user_options.count("parties")) {
    const std::vector<std::string> other_parties{
        user_options["parties"].as<std::vector<std::string>>()};
    if ((other_parties.size() != 3 &&
         (user_options.count("input") || user_options.count("input-file"))) ||
        (other_parties.size() != 2 && user_options.count("input-file-shared")))
      throw std::runtime_error(fmt::format(
          "Incorrect number of parties {} for the chosen input type", other_parties.size()));

    std::string parties("Other parties: ");
    for (auto& party : other_parties) {
      if (CheckPartyArgumentSyntax(party)) {
        if (print) parties.append(" " + party);
      } else {
        throw std::runtime_error(
            fmt::format("Incorrect party argument syntax for party {}", party));
      }
    }
    if (print) std::cout << parties << std::endl;
  } else
    throw std::runtime_error("Other parties' information is not set but required");

  if (!user_options.count("input") && !user_options.count("input-file") &&
      !user_options.count("input-file-shared"))
    throw std::runtime_error("Inputs are not set but required");
  else if ((user_options.count("input") && user_options.count("input-file")) ||
           (user_options.count("input") && user_options.count("input-file-shared")) ||
           (user_options.count("input-file") && user_options.count("input-file-shared")))
    throw std::runtime_error("More than one type of inputs are set but only required one");

  if (print) {
    std::cout << "MPC Protocol: " << user_options["protocol"].as<std::string>() << std::endl;
  }
  return std::make_pair(user_options, std::vector<bool>{help, print_output});
}

encrypto::motion::PartyPointer CreateParty(const program_options::variables_map& user_options) {
  const auto parties_string{user_options["parties"].as<const std::vector<std::string>>()};
  const auto number_of_parties{parties_string.size()};
  const auto my_id{user_options["my-id"].as<std::size_t>()};
  if (my_id >= number_of_parties) {
    throw std::runtime_error(fmt::format(
        "My id needs to be in the range [0, #parties - 1], current my id is {} and #parties is {}",
        my_id, number_of_parties));
  }

  encrypto::motion::communication::TcpPartiesConfiguration parties_configuration(number_of_parties);

  for (const auto& party_string : parties_string) {
    const auto [party_id, host, port] = ParsePartyArgument(party_string);
    if (party_id >= number_of_parties) {
      throw std::runtime_error(
          fmt::format("Party's id needs to be in the range [0, #parties - 1], current id "
                      "is {} and #parties is {}",
                      party_id, number_of_parties));
    }
    parties_configuration.at(party_id) = std::make_pair(host, port);
  }
  encrypto::motion::communication::TcpSetupHelper helper(my_id, parties_configuration);
  auto communication_layer = std::make_unique<encrypto::motion::communication::CommunicationLayer>(
      my_id, helper.SetupConnections());
  auto party = std::make_unique<encrypto::motion::Party>(std::move(communication_layer));
  auto configuration = party->GetConfiguration();
  // disable logging if the corresponding flag was set
  const auto logging{!user_options.count("disable-logging")};
  configuration->SetLoggingEnabled(logging);
  configuration->SetOnlineAfterSetup(user_options["online-after-setup"].as<bool>());
  return party;
}
