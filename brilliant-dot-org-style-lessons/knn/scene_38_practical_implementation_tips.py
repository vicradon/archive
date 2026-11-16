# Streamlit app: show some final tips and a data frame with interactive sorting
import streamlit as st
import pandas as pd
import numpy as np
from sklearn.neighbors import KNeighborsRegressor

st.set_page_config(page_title="Scene 38 — Final Tips", layout="centered")
st.title("Scene 38: Practical Implementation Tips")

st.write("Here's a summary of best practices for successful K-Nearest Neighbors projects:")

st.markdown("""
- **Always standardize numerical features** to prevent a single feature from dominating distance calculations.
- **Start with a small K** and increase gradually to see how it affects your results.
- **Use cross-validation** to choose the best K value.
- **Consider feature selection** to remove noise and reduce the curse of dimensionality.
- **Monitor computational performance** as your data grows, and be ready to use more advanced methods for larger datasets.
- **Validate on a completely unseen test dataset** to get a realistic measure of your model's performance.
""")

st.write("### Example Dataset")

df = pd.DataFrame({
    "Vintage": [1974, 1958, 1969, 1968, 1980],
    "Summer_Temp_C": [16.5, 16.2, 15.7, 16.0, 16.4],
    "Winter_Rain_mm": [760, 610, 690, 700, 730],
    "Price_USD": [8.50, 16.00, 14.50, 15.50, 11.50]
})

st.dataframe(df, use_container_width=True)

st.write("This table shows the 5 nearest neighbors to the 1986 vintage based on weather data.")
