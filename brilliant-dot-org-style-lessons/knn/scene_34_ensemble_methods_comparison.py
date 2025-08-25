# Streamlit app: demonstrate ensemble voting vs. single KNN prediction
import streamlit as st
import pandas as pd
import numpy as np
import plotly.express as px
import plotly.graph_objects as go
from sklearn.neighbors import KNeighborsClassifier
from sklearn.preprocessing import StandardScaler
from sklearn.pipeline import make_pipeline
from sklearn.model_selection import train_test_split
from sklearn.datasets import make_moons
from sklearn.ensemble import VotingClassifier

st.set_page_config(page_title="Scene 34 — Ensemble", layout="centered")
st.title("Scene 34: Ensemble Methods — A Committee of Detectives")
st.write("Combining different K-Nearest Neighbors models for a more robust prediction.")

# Generate a nonlinear, non-trivial classification dataset
X, y = make_moons(n_samples=200, noise=0.2, random_state=42)
df = pd.DataFrame(X, columns=["feature1","feature2"])
df["label"] = y

# Create individual models with different K values
pipe_k1 = make_pipeline(StandardScaler(), KNeighborsClassifier(n_neighbors=1))
pipe_k5 = make_pipeline(StandardScaler(), KNeighborsClassifier(n_neighbors=5))
pipe_k11 = make_pipeline(StandardScaler(), KNeighborsClassifier(n_neighbors=11))

# Create an ensemble (voting classifier)
ensemble = VotingClassifier([('k1', pipe_k1), ('k5', pipe_k5), ('k11', pipe_k11)], voting='hard')

# Fit models on the data
pipe_k1.fit(X, y)
pipe_k5.fit(X, y)
pipe_k11.fit(X, y)
ensemble.fit(X, y)

# Create a meshgrid for plotting decision boundaries
x_min, x_max = X[:, 0].min() - 0.5, X[:, 0].max() + 0.5
y_min, y_max = X[:, 1].min() - 0.5, X[:, 1].max() + 0.5
xx, yy = np.meshgrid(np.arange(x_min, x_max, 0.05), np.arange(y_min, y_max, 0.05))
mesh_data = np.c_[xx.ravel(), yy.ravel()]

# Make predictions on the meshgrid
Z_k1 = pipe_k1.predict(mesh_data).reshape(xx.shape)
Z_k5 = pipe_k5.predict(mesh_data).reshape(xx.shape)
Z_k11 = pipe_k11.predict(mesh_data).reshape(xx.shape)
Z_ensemble = ensemble.predict(mesh_data).reshape(xx.shape)

# Create subplots
fig = go.Figure()
fig.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k1, showscale=False, opacity=0.3, colorscale='Viridis', name='K=1 Prediction'))
fig.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k5, showscale=False, opacity=0.3, colorscale='Viridis', name='K=5 Prediction'))
fig.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k11, showscale=False, opacity=0.3, colorscale='Viridis', name='K=11 Prediction'))
fig.add_trace(go.Scatter(x=X[:, 0], y=X[:, 1], mode='markers', marker=dict(color=y, colorscale='Viridis', symbol='circle', size=8), name='Data Points'))

st.write("### Decision Boundaries for Individual Models")
st.plotly_chart(fig, use_container_width=True)

# Ensemble plot
fig_ensemble = go.Figure()
fig_ensemble.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_ensemble, showscale=False, opacity=0.3, colorscale='Viridis', name='Ensemble Prediction'))
fig_ensemble.add_trace(go.Scatter(x=X[:, 0], y=X[:, 1], mode='markers', marker=dict(color=y, colorscale='Viridis', symbol='circle', size=8), name='Data Points'))

st.write("### Ensemble (Voting) Prediction")
st.plotly_chart(fig_ensemble, use_container_width=True)

st.write("Notice how the ensemble prediction smooths out the noisy, irregular boundaries of the K=1 model and the overly simplified boundaries of the K=11 model, creating a more robust and accurate result.")
