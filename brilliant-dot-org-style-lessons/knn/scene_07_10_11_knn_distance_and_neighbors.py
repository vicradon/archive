# Streamlit app: compute standardized distances + list K nearest neighbors for 1986
import streamlit as st
import pandas as pd
import numpy as np

st.set_page_config(page_title="Scene 7/10/11 — KNN Neighbors", layout="centered")
st.title("Scenes 7, 10 & 11: From Distance to Nearest Neighbors")

# Base data (train known prices, 1986 unknown)
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
    (1986, 16.3, 720, np.nan),  # test
]
df = pd.DataFrame(rows, columns=["year", "temp", "rain", "price"])

# Controls
k = st.slider("K (number of neighbors)", 1, 7, 5, step=1)
use_standardize = st.toggle("Standardize features (recommended)", True)

train = df.dropna(subset=["price"]).copy()
test = df.query("year==1986")[["temp","rain"]].iloc[0].values

X = train[["temp","rain"]].to_numpy()
y = train["price"].to_numpy()

if use_standardize:
    mu = X.mean(axis=0); sigma = X.std(axis=0, ddof=0)
    Xs = (X - mu) / sigma
    ts = (test - mu) / sigma
else:
    Xs = X.copy()
    ts = test.copy()

# Euclidean distances
dists = np.sqrt(((Xs - ts)**2).sum(axis=1))
train = train.assign(distance=dists).sort_values("distance")

st.subheader("Nearest neighbors to 1986 by weather similarity")
st.dataframe(train[["year","temp","rain","price","distance"]].head(k), use_container_width=True)
