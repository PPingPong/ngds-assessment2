"""
Answer key for the Go regression.
Run in the same folder as NGDS_CO2_Algeria.csv:   python check_reference.py
Compare the numbers it prints with the output of `go run .`
"""
import numpy as np
import pandas as pd

CSV = "NGDS_CO2_Algeria.csv"
TARGET = "Flaring Vol (million m3)"
TRAIN_UNTIL = 2020          # same rule as trainUntilYear in the Go code

df = pd.read_csv(CSV)
df.columns = [" ".join(c.split()) for c in df.columns]      # "Field  Type" -> "Field Type"
df["Field Type"] = df["Field Type"].fillna("Unknown")       # blank -> Unknown (baseline)
print(f"Loaded {len(df)} rows")


def design(d):
    """[1, Year, Latitude, Longitude, GAS, LNG, OIL]  (Unknown = baseline)"""
    X = d[["Year", "Latitude", "Longitude"]].astype(float).copy()
    for t in ["GAS", "LNG", "OIL"]:
        X[t] = (d["Field Type"] == t).astype(float)
    return np.column_stack([np.ones(len(d)), X.values])


def fit(d):
    beta, *_ = np.linalg.lstsq(design(d), d[TARGET].values, rcond=None)
    return beta


def metrics(d, beta):
    y = d[TARGET].values
    resid = y - design(d) @ beta
    r2 = 1 - (resid ** 2).sum() / ((y - y.mean()) ** 2).sum()
    return r2, np.sqrt((resid ** 2).mean()), np.abs(resid).mean()


names = ["Intercept", "Year", "Latitude", "Longitude", "GAS", "LNG", "OIL"]
beta = fit(df)
print("\nModel fitted on all rows")
for n, b in zip(names, beta):
    print(f"  {n:<10} {b:12.4f}")
r2, rmse, mae = metrics(df, beta)
print(f"In-sample:  R2 = {r2:.4f}  RMSE = {rmse:.4f}  MAE = {mae:.4f}  (n = {len(df)})")

train, test = df[df.Year <= TRAIN_UNTIL], df[df.Year > TRAIN_UNTIL]
b2 = fit(train)
tr, te = metrics(train, b2), metrics(test, b2)
base = np.sqrt(((test[TARGET] - train[TARGET].mean()) ** 2).mean())
print(f"\nTrain (<= {TRAIN_UNTIL}, n = {len(train)}): R2 = {tr[0]:.4f}  RMSE = {tr[1]:.4f}")
print(f"Test  (>  {TRAIN_UNTIL}, n = {len(test)}): R2 = {te[0]:.4f}  RMSE = {te[1]:.4f}  MAE = {te[2]:.4f}")
print(f"Baseline (predict the training mean): test RMSE = {base:.4f}")
