# Streamlit app: predict 1986 price from K neighbors (with option for distance weighting)
import streamlit as st
import pandas as pd
import numpy as np

st.set_page_config(page_title="Scene 12 — Prediction", layout="centered")
st.title("Scene 12: The Moment of Truth — Predicting 1986")

rows = [
    (1958, 16.2, 610, 16.00),
    (1968, 16.0, 700, 15.50),
    (1969, 15.7, 690, 14.50),
    (1974, 16.5, 760,  8.50),
    (1980, 16.4, 730, 11.50),
    (1978, 16.1, 640, 12.30),
    (1970, 15.8, 650, 12.00),
    (1961, 16.8, 580, 22.00),
    (1955, 15.4, 520, 13.20),
    (1986, 16.3, 720, np.nan),
]
df = pd.DataFrame(rows, columns=["year","temp","rain","price"])
train = df.dropna(subset=["price"])
x_test = df.query("year==1986")[["temp","rain"]].iloc[0].values

k = st.slider("K", 1, 7, 5)
standardize = st.toggle("Standardize features", True)
weighted = st.toggle("Weighted KNN (1/distance)", False)

X = train[["temp","rain"]].to_numpy()
y = train["price"].to_numpy()

if standardize:
    mu, sd = X.mean(axis=0), X.std(axis=0, ddof=0)
    Xs = (X - mu)/sd
    ts = (x_test - mu)/sd
else:
    Xs = X.copy(); ts = x_test.copy()

d = np.sqrt(((Xs - ts)**2).sum(axis=1))
order = np.argsort(d)[:k]
neighbors = train.iloc[order].assign(distance=d[order])

if weighted:
    # avoid division by zero
    w = 1/np.maximum(neighbors["distance"].to_numpy(), 1e-9)
    pred = (neighbors["price"].to_numpy()*w).sum()/w.sum()
else:
    pred = neighbors["price"].mean()

st.metric("Predicted 1986 price (USD)", f"")
st.caption("This uses the selected K, scaling, and weighting options above.")
st.dataframe(neighbors[["year","temp","rain","price","distance"]], use_container_width=True)
