# Streamlit app: weather vs. price scatter with interactive hover
import streamlit as st
import pandas as pd
import numpy as np
import plotly.express as px

st.set_page_config(page_title="Scene 4 — Weather Map", layout="centered")

st.title("Scene 4: The Visualization Reveals All")
st.write("Winter rainfall vs. summer temperature, colored by wine price.")

# Toy dataset with story-relevant years included
data = [
    # year, temp, rain, price
    (1958, 16.2, 610, 16.00),
    (1968, 16.0, 700, 15.50),
    (1969, 15.7, 690, 14.50),
    (1974, 16.5, 760,  8.50),
    (1980, 16.4, 730, 11.50),
    (1986, 16.3, 720, np.nan),  # test year (unknown price)
    # a few filler points
    (1955, 15.4, 520, 13.20),
    (1961, 16.8, 580, 22.00),
    (1970, 15.8, 650, 12.00),
    (1975, 16.6, 770,  9.80),
    (1978, 16.1, 640, 12.30),
]
df = pd.DataFrame(data, columns=["year", "summer_temp_c", "winter_rain_mm", "price_usd"])

highlight_1986 = st.toggle("Highlight 1986", True)
fig = px.scatter(
    df, x="winter_rain_mm", y="summer_temp_c",
    color="price_usd", hover_data=["year", "price_usd"],
    color_continuous_scale="Viridis",
    title="Weather space (rain vs temp), colored by price"
)

if highlight_1986:
    y1986 = df.query("year==1986").iloc[0]
    fig.add_scatter(
        x=[y1986["winter_rain_mm"]], y=[y1986["summer_temp_c"]],
        mode="markers+text", text=["1986"], textposition="top center",
        marker=dict(size=14, line=dict(width=2, color="black"))
    )

st.plotly_chart(fig, use_container_width=True)
