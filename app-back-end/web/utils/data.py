import hashlib
import scipy.stats as stats


def get_sample_size(confidence_level, error, total_sample_size, P=0.5):
    alpha = 1 - confidence_level
    z_alpha_2 = stats.norm.ppf(1 - alpha / 2)
    P_one_minus_P = P * (1 - P)

    if total_sample_size >= 1_0000_0000:
        n_infinite = (z_alpha_2 ** 2 * P_one_minus_P) / (error ** 2)
        return int(n_infinite)
    else:
        n_finite = (z_alpha_2 ** 2 * P_one_minus_P * total_sample_size) / ((total_sample_size - 1) * error ** 2 + z_alpha_2 ** 2 * P_one_minus_P)
        return int(n_finite)
    

async def data_summary(file_cont: bytes):
    md5_hash = hashlib.md5()
    md5_hash.update(file_cont)
    file_md5 = md5_hash.hexdigest()

    df = pd.read_csv(BytesIO(file_cont))
    if "data" not in df.columns:
        raise ValueError("The column 'data' is missing in the file.")

    return {
        "md5": file_md5,
        "items": len(df),
        "mean": round(float(df["data"].mean()), 7),
        "std": round(float(df["data"].std()), 7),
        "max": round(float(df["data"].max()), 7),
        "min": round(float(df["data"].min()), 7)
    }