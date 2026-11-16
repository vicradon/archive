# Streamlit app: visualize neighbor contributions under distance weighting
import streamlit as st
import pandas as pd
import numpy as np
import plotly.express as px

st.set_page_config(page_title="Scene 30 — Weighted KNN", layout="centered")
st.title("Scene 30: Weighted KNN — Who Influences the Prediction?")

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
test = df.query("year==1986")[["temp","rain"]].iloc[0].to_numpy()

k = st.slider("K", 1, 7, 5)
standardize = st.toggle("Standardize features", True)
power = st.slider("Weight falloff exponent (1/d^p)", 0.5, 3.0, 1.0, 0.1)

X = train[["temp","rain"]].to_numpy()
y = train["price"].to_numpy()

if standardize:
    mu, sd = X.mean(axis=0), X.std(axis=0, ddof=0)
    X = (X - mu)/sd
    t = (test - mu)/sd
else:
    t = test.copy()

dist = np.sqrt(((X - t)**2).sum(axis=1))
order = np.argsort(dist)[:k]
nn = train.iloc[order].copy()

eps = 1e-9
weights = 1/np.maximum(dist[order], eps)**power
weights = weights/weights.sum()

nn["weight"] = weights
nn["contribution"] = nn["price"] * nn["weight"]
pred = nn["contribution"].sum()

st.metric("Weighted KNN Prediction", f"")
fig = px.bar(nn, x="year", y="contribution", hover_data=["price","weight"], title="Neighbor contributions to prediction")
st.plotly_chart(fig, use_container_width=True)
st.dataframe(nn[["year","price","weight","contribution"]], use_container_width=True)
