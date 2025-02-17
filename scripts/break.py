#!/home/b426/miniconda3/envs/mpc/bin/python

import argparse
import pandas as pd
import dask.dataframe as dd

from dask.diagnostics import ProgressBar


def do_break(filename, break_rate):
    df = pd.read_hdf(filename, key='data')

    if break_rate < 0 or break_rate > 1:
        raise ValueError("break_rate should between 0 and 1.")
    
    num_rows_to_break = int(len(df) * break_rate)
    df.loc[:num_rows_to_break - 1, 'data'] = 0.0
    
    ddf = dd.from_pandas(df)
    pbar = ProgressBar()
    pbar.register()
    
    ddf.to_hdf(filename, key='data', mode='w')
    pbar.unregister()


def main():
    parser = argparse.ArgumentParser(description="Break the file")
    parser.add_argument('-f', '--file', required=True, help="File to break")
    parser.add_argument('-b', '--break-rate', type=float, default=0.1, help="Break rate")
    args = parser.parse_args()

    import time
    start = time.time()
    do_break(args.file, args.break_rate)
    print(f"Breaked {args.break_rate * 100}% of the data in file {args.file}.")
    print(f"Time cost: {time.time() - start}")


if __name__ == '__main__':
    main()
