# Streamlit app: compare distance metrics (Euclidean vs Manhattan vs Cosine) on prediction
import streamlit as st
import pandas as pd
import numpy as np
from sklearn.preprocessing import StandardScaler
from numpy.linalg import norm

st.set_page_config(page_title="Scene 24 — Distance Metrics", layout="centered")
st.title("Scene 24: Distance Metrics Beyond Euclidean")

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
train = df.dropna(subset=["price"]).copy()
x_test = df.query("year==1986")[["temp","rain"]].to_numpy()

k = st.slider("K", 1, 7, 5)
metric = st.selectbox("Distance metric", ["euclidean","manhattan","cosine"])
standardize = st.toggle("Standardize features", True)

X = train[["temp","rain"]].to_numpy()
y = train["price"].to_numpy()
Xt = x_test.copy()

if standardize:
    scaler = StandardScaler().fit(X)
    X = scaler.transform(X)
    Xt = scaler.transform(Xt)

def pdist(A, b, metric):
    if metric == "euclidean":
        return np.sqrt(((A - b)**2).sum(axis=1))
    if metric == "manhattan":
        return np.abs(A - b).sum(axis=1)
    if metric == "cosine":
        # 1 - cosine similarity
        return 1 - (A @ b.T).ravel()/(norm(A,axis=1)*norm(b))
    raise ValueError(metric)

d = pdist(X, Xt, metric)
idx = np.argsort(d)[:k]
pred = y[idx].mean()

st.metric("Prediction for 1986", f"")
st.dataframe(train.iloc[idx].assign(distance=d[idx])[["year","temp","rain","price","distance"]], use_container_width=True)
