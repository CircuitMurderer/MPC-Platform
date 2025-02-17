import dask.dataframe as dd
import dask.array as da
import argparse

from dask.diagnostics import ProgressBar


def generate_csv(filename, N, is_csv):
    df = dd.from_array(da.arange(1, N + 1), columns=['number'])
    df['data'] = da.random.uniform(0, 1, size=N)
    df = df.set_index('number')

    print('Writing...')
    pbar = ProgressBar()
    pbar.register()

    if is_csv:
        df.to_csv(filename, index=True, index_label='number')
    else:
        df.to_hdf(filename, key='data', mode='w')

    pbar.unregister()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate data with random data from 0 to 1.")
    parser.add_argument('-n', '--number', type=int, required=True, help="Number of rows to generate.")
    parser.add_argument('-f', '--filename', type=str, required=True, help="Output filename.")
    parser.add_argument('--csv', action='store_true', default=False, help="Pure CSV mode.")
    args = parser.parse_args()
    
    import time
    start = time.time()
    generate_csv(args.filename, args.number, args.csv)

    print(f"Generated {args.number} rows of data in {args.filename}.")
    print(f"Time cost: {time.time() - start}")
