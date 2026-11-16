# Streamlit app: final summary and interactive data frame for reflection
import streamlit as st
import pandas as pd
import numpy as np
from sklearn.preprocessing import StandardScaler
from sklearn.neighbors import KNeighborsRegressor

st.set_page_config(page_title="Scene 40 — The Story Continues", layout="centered")
st.title("Scene 40: The Continuing Detective Story")

st.write("Our journey through K-Nearest Neighbors teaches us that models are tools, not crystal balls. The 1986 wine prediction failed, but it revealed invaluable lessons about the limitations of data and the importance of human judgment.")

st.write("### Key Takeaways")
st.markdown("""
- **Similarity-based reasoning** is intuitive and powerful.
- **Feature scaling** is crucial for fair distance calculations.
- **Model validation** helps us choose optimal parameters.
- **Interpretability** makes models trustworthy and actionable.
- **Failures teach us** about model limitations and the need for human oversight.
""")

st.write("### Revisit the Prediction")
st.write("You can adjust the parameters below to see how our prediction for the 1986 vintage changes.")

# Load the data again
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
    (1981, 16.0, 680, np.nan),
    (1982, 16.2, 650, np.nan),
    (1983, 15.9, 690, np.nan),
    (1984, 16.4, 710, np.nan),
]
df = pd.DataFrame(rows, columns=["year","temp","rain","price"])
train = df.dropna(subset=["price"])
test = df[df["price"].isna()]

# Controls
k = st.slider("K", 1, 9, 5)
standardize = st.toggle("Standardize features", True)
weighted = st.toggle("Weighted KNN", False)

# Run the model
if standardize:
    scaler = StandardScaler().fit(train[["temp","rain"]])
    X_train_scaled = scaler.transform(train[["temp","rain"]])
    X_test_scaled = scaler.transform(test[["temp","rain"]])
else:
    X_train_scaled = train[["temp","rain"]]
    X_test_scaled = test[["temp","rain"]]

model = KNeighborsRegressor(n_neighbors=k, weights="distance" if weighted else "uniform")
model.fit(X_train_scaled, train["price"])
test["prediction"] = model.predict(X_test_scaled)

st.dataframe(test[["year","temp","rain","prediction"]], use_container_width=True)

st.write("The official price of the 1986 vintage was later revealed to be 5.00, proving the limitations of the model. However, the process of using it to make predictions and see its limitations is the valuable lesson.")
