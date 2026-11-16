# Streamlit app: sklearn Pipeline + cross-validation over K with RMSE curve
import streamlit as st
import pandas as pd
import numpy as np
import plotly.express as px
from sklearn.model_selection import KFold, cross_val_score
from sklearn.pipeline import make_pipeline
from sklearn.preprocessing import StandardScaler
from sklearn.neighbors import KNeighborsRegressor

st.set_page_config(page_title="Scene 15 — Pipelines & CV", layout="centered")
st.title("Scene 15: Pipelines + Cross-Validation")

# slightly larger synthetic dataset for demo
rng = np.random.default_rng(42)
N = st.slider("Dataset size", 40, 400, 120, step=20)
temp = rng.normal(16.2, 0.5, N)
rain = rng.normal(670, 70, N)
# latent relation + noise (purely for demo)
price = 5 + 2.2*(temp-16) - 0.004*(rain-670) + rng.normal(0, 1.2, N)

df = pd.DataFrame({"temp":temp, "rain":rain, "price":price})

k_values = list(range(1, 26, 2))
cv_folds = st.slider("CV folds", 3, 10, 5)

rmse = []
for k in k_values:
    pipe = make_pipeline(StandardScaler(), KNeighborsRegressor(n_neighbors=k))
    # negative MSE -> convert to RMSE
    scores = cross_val_score(pipe, df[["temp","rain"]], df["price"], scoring="neg_mean_squared_error", cv=KFold(cv_folds, shuffle=True, random_state=0))
    rmse.append(np.sqrt(-scores).mean())

cv_df = pd.DataFrame({"K":k_values,"RMSE":rmse})
fig = px.line(cv_df, x="K", y="RMSE", markers=True, title="CV RMSE vs K (lower is better)")
st.plotly_chart(fig, use_container_width=True)

best = cv_df.loc[cv_df["RMSE"].idxmin()]
st.success(f"Best K ≈ {int(best.K)} with RMSE ≈ {best.RMSE:.3f}")
