require 'open3'

def run_program(command)
  stdout, stderr, status = Open3.capture3 command
  { 
    stdout: stdout, 
    stderr: stderr, 
    status: status.exitstatus 
  }
end

cal_cmd = 'build/mult3'
cal_args = {
  '--protocol': 'BMR', # ArithmeticGMW, BooleanGMW, BMR
  '--parties': [
      '0,127.0.0.1,23000',
      '1,127.0.0.1,23001',
      '2,127.0.0.1,23002'
  ].join(' '),
  '--print-output': nil,
  '--my-id': [0, 1, 2],
  '--input': [6, 3, 8]
}

t_cmd = ['', '', '']
(0...3).each do |i|
  t_cmd[i] = cal_cmd.dup
  cal_args.each do |k, v|
    t_cmd[i] << ' ' << "#{k}" if v.nil?
    t_cmd[i] << ' ' << "#{k}=#{v}" if v.is_a? String 
    t_cmd[i] << ' ' << "#{k}=#{v[i]}" if v.is_a? Array 
  end
end

t1 = Thread.new { run_program t_cmd[0] }
t2 = Thread.new { run_program t_cmd[1] }
t3 = Thread.new { run_program t_cmd[2] }

result1 = t1.value
result2 = t2.value
result3 = t3.value

puts "Output of x1:"
puts result1[:stdout]
puts "Errors of x1:" unless result1[:stderr].empty?
puts result1[:stderr] unless result1[:stderr].empty?
puts "Exit status of x1: #{result1[:status]}"

puts "Output of x2:"
puts result2[:stdout]
puts "Errors of x2:" unless result2[:stderr].empty?
puts result2[:stderr] unless result2[:stderr].empty?
puts "Exit status of x2: #{result2[:status]}"

puts "Output of x3:"
puts result3[:stdout]
puts "Errors of x3:" unless result3[:stderr].empty?
puts result3[:stderr] unless result3[:stderr].empty?
puts "Exit status of x3: #{result3[:status]}"
