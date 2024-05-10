import random


def write_to_file(data, filename):
    with open(filename, 'w', encoding='utf-8') as f:
        for i, v in enumerate(data):
            f.write(str(v))
            if (i + 1) % 16 == 0:
                f.write('\n')
            else:
                f.write(' ')


rand_max = 1_000_000
rand_len = [1000, 10_000, 100_000, 1_000_000, 10_000_000]



for l in rand_len:
    a_s, b_s, add_s, mul_s = [], [], [], []
    for _ in range(l):
        a = random.randint(1, rand_max)
        b = random.randint(1, rand_max)

        add_res = a + b
        mul_res = a * b

        a_s.append(a)
        b_s.append(b)
        add_s.append(add_res)
        mul_s.append(mul_res)

    write_to_file(a_s, f'./bench_datas/{l}_A.txt')
    write_to_file(b_s, f'./bench_datas/{l}_B.txt')
    write_to_file(add_s, f'./bench_datas/{l}_Add.txt')
    write_to_file(mul_s, f'./bench_datas/{l}_Mul.txt')

    

    
