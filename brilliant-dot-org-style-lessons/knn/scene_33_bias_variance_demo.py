# Streamlit app: show bias/variance trade-off by varying K
import streamlit as st
import pandas as pd
import numpy as np
import plotly.graph_objects as go
from sklearn.neighbors import KNeighborsRegressor
from sklearn.preprocessing import StandardScaler
from sklearn.pipeline import make_pipeline

st.set_page_config(page_title="Scene 33 — Bias/Variance", layout="centered")
st.title("Scene 33: The Bias-Variance Tradeoff")
st.write("How K influences model complexity (bias) and sensitivity (variance).")

# Simulate a nonlinear data generating process
def get_data(n=200):
    rng = np.random.default_rng(42)
    X = np.linspace(0, 10, n).reshape(-1, 1)
    y = np.sin(X.ravel()) + 0.3 * np.cos(5*X.ravel()) + rng.normal(0, 0.2, n)
    return X, y

X_train, y_train = get_data(50)
X_test, y_test = get_data(200)

k = st.slider("K (number of neighbors)", 1, 30, 1)

# Fit model
pipe = make_pipeline(StandardScaler(), KNeighborsRegressor(n_neighbors=k))
pipe.fit(X_train, y_train)

# Make predictions on a fine grid
X_pred = np.linspace(0, 10, 1000).reshape(-1, 1)
y_pred = pipe.predict(X_pred)

# Plotting
fig = go.Figure()
fig.add_trace(go.Scatter(x=X_train.ravel(), y=y_train, mode='markers', name='Training Data'))
fig.add_trace(go.Scatter(x=X_test.ravel(), y=y_test, mode='markers', name='Test Data (true function)'))
fig.add_trace(go.Scatter(x=X_pred.ravel(), y=y_pred, mode='lines', name=f'Prediction (K={k})', line=dict(color='red', width=3)))
fig.update_layout(title=f"K={k} KNN Prediction", xaxis_title="X", yaxis_title="Y")

st.plotly_chart(fig, use_container_width=True)

st.write("### Analysis")
st.write(f"- **Low K (e.g., K=1):** The prediction line wiggles to fit every single training point. This is **low bias** (it's flexible) but **high variance** (it's sensitive to noise in the training data).")
st.write(f"- **High K (e.g., K=30):** The prediction line becomes much smoother, ignoring individual data points. This is **low variance** (it's stable) but **high bias** (it might miss the true, complex pattern).")
