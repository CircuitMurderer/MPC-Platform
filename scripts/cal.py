import dask.dataframe as dd
import dask.array as da
import argparse

from dask.diagnostics import ProgressBar


def do_calculate(file_a, file_b, file_r, operator):
    df_a = dd.read_hdf(file_a, key='data')
    df_b = dd.read_hdf(file_b, key='data')

    pbar = ProgressBar()
    print('Init...')
    pbar.register()

    data_a = da.from_array(df_a['data'].compute())
    data_b = da.from_array(df_b['data'].compute())

    df_r = dd.from_array(da.arange(1, data_a.shape[0] + 1), columns=['number'])
    df_r = df_r.set_index('number')

    pbar.unregister()
    print('Calculating...', data_a.shape, data_b.shape)

    if operator.lower() == 'add':
        data_r = data_a + data_b
    elif operator.lower() == 'sub':
        data_r = data_a - data_b
    elif operator.lower() == 'mul':
        data_r = data_a * data_b
    elif operator.lower() == 'div':
        data_r = data_a / data_b
    elif operator.lower() == 'exp':
        data_r = data_a ** data_b
    else:
        raise ValueError(f"Unsupported operator: {operator}")

    pbar.register()
    df_r['data'] = data_r.compute()
    pbar.unregister()

    print('Saving...')
    pbar.register()
    df_r.to_hdf(file_r, key='data', mode='w')
    pbar.unregister()


def main():
    parser = argparse.ArgumentParser(description="Perform operations on two CSV files.")
    parser.add_argument('-a', '--file-a', required=True, help="First input CSV file.")
    parser.add_argument('-b', '--file-b', required=True, help="Second input CSV file.")
    parser.add_argument('-o', '--operator', required=True, help="Operator to apply.")
    parser.add_argument('-f', '--file-out', default='result.csv', help="Output CSV file.")
    args = parser.parse_args()

    import time
    start = time.time()
    do_calculate(args.file_a, args.file_b, args.file_out, args.operator)
    print(f"Calculation with operator '{args.operator}' completed successfully! Output saved to {args.file_out}.")
    print(f"Time cost: {time.time() - start}")


if __name__ == '__main__':
    main()
