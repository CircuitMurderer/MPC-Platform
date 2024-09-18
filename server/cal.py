import csv
import argparse


def read_csv_data(filename):
    data = []
    with open(filename, 'r') as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            data.append(float(row['data']))
    return data


def write_csv_data(filename, data):
    with open(filename, 'w', newline='') as csvfile:
        fieldnames = ['number', 'data']
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)

        writer.writeheader()
        for i, value in enumerate(data, start=1):
            writer.writerow({'number': i, 'data': value})


def apply_operation(data1, data2, operator: str):
    if operator.lower() == 'add':
        return [d1 + d2 for d1, d2 in zip(data1, data2)]
    elif operator.lower() == 'sub':
        return [d1 - d2 for d1, d2 in zip(data1, data2)]
    elif operator.lower() == 'mul':
        return [d1 * d2 for d1, d2 in zip(data1, data2)]
    elif operator.lower() == 'div':
        return [d1 / d2 if d2 != 0 else float('inf') for d1, d2 in zip(data1, data2)]
    elif operator.lower() == 'exp':
        return [d1 ** d2 for d1, d2 in zip(data1, data2)]
    else:
        raise ValueError(f"Unsupported operator: {operator}")


def main():
    parser = argparse.ArgumentParser(description="Perform operations on two CSV files.")
    parser.add_argument('-a', '--file-a', required=True, help="First input CSV file.")
    parser.add_argument('-b', '--file-b', required=True, help="Second input CSV file.")
    parser.add_argument('-o', '--operator', required=True, help="Operator to apply.")
    parser.add_argument('-f', '--file-out', default='result.csv', help="Output CSV file.")
    args = parser.parse_args()

    data1 = read_csv_data(args.file_a)
    data2 = read_csv_data(args.file_b)

    if len(data1) != len(data2):
        print("The two CSV files have different lengths.")
        return

    result_data = apply_operation(data1, data2, args.operator)
    write_csv_data(args.file_out, result_data)
    print(f"Calculation with operator '{args.operator}' completed successfully! Output saved to {args.file_out}.")


if __name__ == '__main__':
    main()

