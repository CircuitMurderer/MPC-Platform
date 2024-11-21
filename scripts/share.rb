require 'open3'

def run_program(command)
  stdout, stderr, status = Open3.capture3 command
  { 
    stdout: stdout, 
    stderr: stderr, 
    status: status.exitstatus 
  }
end

g_args = {
  port: 8001,
  address: '127.0.0.1',
  csvfile: 'data10k.csv',
  sharefile: 'Share.bin',
  basepath: 'data/',
}

cal_cmd = 'build/sharer'
cal_args = [
  {
    'ro': 1,
    'pt': g_args[:port],
    'csv': g_args[:csvfile],
    'shr': g_args[:sharefile],
    'pth': g_args[:basepath],
  },
  {
    'ro': 2,
    'pt': g_args[:port],
    'csv': g_args[:csvfile],
    'shr': g_args[:sharefile],
    'pth': g_args[:basepath],
  }
]

t_cmd = ['', '']
(0...2).each do |i|
  t_cmd[i] = cal_cmd.dup
  cal_args[i].each do |k, v|
    t_cmd[i] << ' ' << "#{k}=#{v}"
  end
end

puts t_cmd

t1 = Thread.new { run_program t_cmd[0] }
t2 = Thread.new { run_program t_cmd[1] }

result1 = t1.value
result2 = t2.value

puts "\033[32mOutput of Alice:\033[0m"
puts result1[:stdout]
puts "\033[31mErrors of Alice:\033[0m" unless result1[:stderr].empty?
puts result1[:stderr] unless result1[:stderr].empty?
puts "\033[33mExit status of Alice:\033[0m #{result1[:status]}"

puts "\033[32mOutput of Bob:\033[0m"
puts result2[:stdout]
puts "\033[31mErrors of Bob:\033[0m" unless result2[:stderr].empty?
puts result2[:stderr] unless result2[:stderr].empty?
puts "\033[33mExit status of Bob:\033[0m #{result2[:status]}"

