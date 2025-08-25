import streamlit as st
import pandas as pd
import numpy as np
import plotly.express as px
import plotly.graph_objects as go
from sklearn.model_selection import KFold, cross_val_score, train_test_split
from sklearn.pipeline import make_pipeline
from sklearn.preprocessing import StandardScaler
from sklearn.neighbors import KNeighborsRegressor, KNeighborsClassifier
from sklearn.metrics import mean_squared_error
from numpy.linalg import norm
from sklearn.impute import KNNImputer
from sklearn.datasets import make_moons
from sklearn.ensemble import VotingClassifier

# --- Streamlit App Configuration ---
st.set_page_config(page_title="The Wine Detective: K-Nearest Neighbors", layout="wide")

# --- Helper Function to Process App Code for Embedding ---
def process_code_block(code_str, scene_title):
    # Remove st.set_page_config as it can only be called once per app
    processed_code = [line for line in code_str.splitlines() if "st.set_page_config" not in line]

    # Adjust st.title to st.subheader and add unique keys to widgets
    processed_lines = []
    # Generate a unique prefix for keys based on the scene title
    key_prefix = "s_" + "_".join(scene_title.lower().split(" "))[:30].replace(".", "").replace(",", "").replace("/", "") # Sanitize key
    
    widget_counter = {} # To ensure unique keys within a single scene's app_code
    
    for line in processed_code:
        if line.strip().startswith("st.title("):
            # Replace st.title with st.subheader for better integration
            processed_lines.append(f'st.subheader("Interactive Demo: {scene_title.split(":", 1)[-1].strip()}")')
        elif any(widget_type in line for widget_type in ["st.slider(", "st.toggle(", "st.selectbox(", "st.radio(", "st.file_uploader("]):
            if "key=" not in line:
                # Determine widget type and assign a counter
                widget_type_key = ""
                if "st.slider(" in line: widget_type_key = "slider"
                elif "st.toggle(" in line: widget_type_key = "toggle"
                elif "st.selectbox(" in line: widget_type_key = "selectbox"
                elif "st.radio(" in line: widget_type_key = "radio"
                elif "st.file_uploader(" in line: widget_type_key = "uploader"
                
                widget_counter[widget_type_key] = widget_counter.get(widget_type_key, 0) + 1
                unique_key = f"{key_prefix}_{widget_type_key}_{widget_counter[widget_type_key]}"
                
                # Insert the key argument before the closing parenthesis
                if line.strip().endswith(")"):
                    line = line[:-1] + f', key="{unique_key}")'
                else: # Handle multiline widget definitions if necessary (though less common for these simple cases)
                    line += f' key="{unique_key}"'
            processed_lines.append(line)
        else:
            processed_lines.append(line)
    return "\n".join(processed_lines)

# --- Define all Scenes with their Markdown and (optional) App Code ---
scenes = {
    "intro": {
        "title": "Introduction",
        "markdown": """
        # The Wine Detective: Cracking the Code of K-Nearest Neighbors

        ## Scene 1: A Mysterious Vintage
        *The year is 1986. Wine collectors across France are buzzing with anticipation.*

        Picture this: You're standing in the vineyards of Bordeaux, watching the last grapes of the 1986 harvest being picked. The summer was warm, the winter rainfall moderate. But here's the million-dollar question that has everyone on edge:

        **Will this vintage be worth collecting, or should investors steer clear?**

        Traditional wine experts would need years to taste and evaluate the wine. But Professor Orley Ashenfelter had a radical idea: What if we could predict wine quality using just weather data?

        ---

        ## Scene 2: The Pattern Detective's Dilemma
        *Our story begins with a simple observation...*

        Ashenfelter noticed something intriguing: wines from years with similar weather patterns often had similar quality and prices. It was like nature left fingerprints on each vintage.

        Think about it — if you wanted to predict how good an unknown wine might be, what would be your strategy?

        **Quiz Time!**
        If you had to predict the quality of a mystery wine, which approach sounds most reasonable?

        A) Flip a coin and hope for the best
        B) Look at wines from years with similar weather and see how they turned out
        C) Ask a magic 8-ball

        *Take a moment to think about it...*

        ✅ **Answer: B** — This is exactly the intuition behind our detective story! We're going to look for similar \"neighbors\" in the data.

        ---

        ## Scene 3: Meet the Data
        *Every good detective story needs evidence...*

        Here's what Ashenfelter collected for each vintage year:
        - **Summer temperature** (how warm it got)
        - **Winter rainfall** (how much it rained)
        - **Wine price** years later (our measure of quality)

        But wait — there's a twist! For the years 1981-1991, he had the weather data but not the wine prices yet. These wines were still too young to evaluate.

        **The mission:** Use the older vintages (1952-1980) to predict the prices of the newer ones.
        """
    },
    "scene_04_visualization_reveals_all": {
        "title": "Scene 4: The Visualization Reveals All",
        "markdown": """
        ## Scene 4: The Visualization Reveals All
        *Let's see what patterns emerge when we plot the data...*

        When we create a scatter plot with winter rainfall on one axis and summer temperature on the other, something fascinating happens. Each dot represents a vintage year, colored by wine price:

        - 🟡 Yellow dots = expensive, high-quality wines
        - 🔵 Blue dots = cheaper, lower-quality wines

        Now here's where it gets interesting: Notice how wines with similar weather conditions (dots close together) tend to have similar colors?
        """,
        "app_code": process_code_block("""
            # Streamlit app: weather vs. price scatter with interactive hover
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
        """, "Scene 4: The Visualization Reveals All")
    },
    "scene_05_the_1986_mystery": {
        "title": "Scene 5: The 1986 Mystery",
        "markdown": """
        ## Scene 5: The 1986 Mystery
        *Our plot thickens...*

        Looking at our weather map, we can see exactly where 1986 lands. The key insight that started our whole investigation:

        **The wines closest to 1986 in weather conditions were mostly low-quality vintages!**

        This is the \"aha!\" moment. The nearest neighbors of 1986 were pointing toward disappointment.

        **Quick Reflection:** Why do you think looking at \"nearest neighbors\" might be a good prediction strategy?
        """
    },
    "scene_06_formalizing_our_detective_work": {
        "title": "Scene 6: Formalizing Our Detective Work",
        "markdown": """
        ## Scene 6: Formalizing Our Detective Work
        *Time to turn intuition into a systematic method...*

        Our wine detective approach has two key ingredients:

        1.  **Training data** — vintages where we know both weather AND final wine quality
        2.  **Test data** — vintages where we only know the weather (what we want to predict)

        This is the fundamental setup of machine learning: learn from examples where you know the answer, then apply that knowledge to new cases.
        """
    },
    "scene_07_10_11_knn_distance_and_neighbors": {
        "title": "Scene 7, 10 & 11: KNN Distance & Neighbors",
        "markdown": """
        ## Scene 7: The K-Nearest Neighbors Algorithm Unveiled
        *Here's how our detective method works step by step...*

        **The K-Nearest Neighbors Algorithm:**

        **Step 1:** For a new wine (like 1986), find the K \"nearest\" wines in the training data based on weather similarity.

        **Step 2:** Average the prices of these K nearest neighbors.

        **Step 3:** That average becomes your prediction!

        The \"K\" is just a number we choose — how many neighbors should we consult? 3? 5? 10?

        ---

        ## Quiz Break: Understanding Distance

        Before we continue, let's test your understanding:

        **What makes two wines \"nearest neighbors\"?**

        A) They were made in consecutive years
        B) They have similar weather patterns
        C) They're from the same vineyard

        ✅ **Answer: B** — We measure \"nearness\" by how similar the weather conditions were, not by time or location!

        ---

        ## Scene 8: The Importance of Fair Comparison
        *But wait — we have a problem...*

        Imagine comparing wines where:
        - Summer temperature ranges from 15°C to 18°C (small range)
        - Winter rainfall ranges from 400mm to 800mm (huge range)

        If we calculate distance directly, rainfall will dominate simply because its numbers are bigger. It's like comparing apples measured in grams to oranges measured in kilograms!

        **The solution:** Scale both features to have similar ranges before calculating distances.

        ---

        ## Scene 9: Scaling the Data
        *Making fair comparisons...*

        **Standardization** transforms each feature to have:
        - Mean = 0
        - Standard deviation = 1

        Think of it as converting everything to the same \"units\" so no feature unfairly dominates the distance calculation.

        **Mathematical note:** For each feature, we calculate:
        `scaled_value = (original_value - mean) / standard_deviation`

        ---

        ## Scene 10: Calculating Distances
        *Now for some mathematical detective work...*

        To find the nearest neighbors to 1986, we calculate the **Euclidean distance** between 1986 and every wine in our training data.

        For two wines with weather conditions (x₁, y₁) and (x₂, y₂):
        **Distance = √[(x₁-x₂)² + (y₁-y₂)²]**

        This gives us a single number representing how \"far apart\" two wines are in weather space.

        ---

        ## Quiz Break: Distance Intuition

        **Quick check:** If two wines have identical weather conditions, what will their distance be?

        A) 1
        B) 0
        C) Impossible to tell

        ✅ **Answer: B** — Identical conditions mean zero distance between them!

        ---

        ## Scene 11: The Nearest Neighbors of 1986
        *Our investigation gets specific...*

        After calculating all distances and sorting them, the 5 nearest neighbors to 1986 are:
        - 1974
        - 1958
        - 1969
        - 1968
        - 1980

        These were the years with the most similar weather patterns to 1986.
        """,
        "app_code": process_code_block("""
            # Streamlit app: compute standardized distances + list K nearest neighbors for 1986
            st.write("This app demonstrates how distances are calculated and nearest neighbors are identified for the 1986 vintage.")

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
        """, "Scene 7, 10 & 11: KNN Distance & Neighbors")
    },
    "scene_12_making_the_prediction": {
        "title": "Scene 12: Making the Prediction",
        "markdown": """
        ## Scene 12: Making the Prediction
        *The moment of truth...*

        The prices of our 5 nearest neighbors were:
        - 1974: $8.50
        - 1958: $16.00
        - 1969: $14.50
        - 1968: $15.50
        - 1980: $11.50

        **Average = (8.50 + 16.00 + 14.50 + 15.50 + 11.50) ÷ 5 = $13.20**

        Our algorithm predicts 1986 will be worth $13.20 — suggesting a mediocre vintage.
        """,
        "app_code": process_code_block("""
            # Streamlit app: predict 1986 price from K neighbors (with option for distance weighting)
            st.write("Adjust K, standardization, and weighting to see how the prediction for 1986 changes.")

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

            st.metric("Predicted 1986 price (USD)", f"${pred:,.2f}")
            st.caption("This uses the selected K, scaling, and weighting options above.")
            st.dataframe(neighbors[["year","temp","rain","price","distance"]], use_container_width=True)
        """, "Scene 12: Making the Prediction")
    },
    "quiz_break_algorithm_understanding": {
        "title": "Quiz Break: Algorithm Understanding",
        "markdown": """
        ## Quiz Break: Algorithm Understanding

        **If we had used K=3 instead of K=5, would our prediction be:**

        A) Exactly the same
        B) Probably different
        C) More accurate

        ✅ **Answer: B** — Different K values use different numbers of neighbors, leading to different predictions. The \"best\" K depends on the specific dataset.
        """
    },
    "scene_13_implementing_from_scratch": {
        "title": "Scene 13: Implementing from Scratch",
        "markdown": """
        ## Scene 13: Implementing from Scratch
        *Let's code our detective work...*

        Here's how we'd implement this by hand:

        ```python
        # Step 1: Scale the features
        X_train_scaled = (X_train - X_train.mean()) / X_train.std()
        X_test_scaled = (X_test - X_train.mean()) / X_train.std()

        # Step 2: Calculate distances
        distances = []
        for each wine in training_data:
            distance = sqrt(sum((test_wine - train_wine)**2))
            distances.append(distance)

        # Step 3: Find K nearest neighbors
        nearest_indices = sort(distances)[:K]

        # Step 4: Average their prices
        prediction = mean(prices[nearest_indices])
        ```
        """
    },
    "scene_14_the_professional_detective_tools": {
        "title": "Scene 14: The Professional Detective Tools",
        "markdown": """
        ## Scene 14: The Professional Detective Tools
        *Enter Scikit-Learn...*

        Rather than implement everything from scratch, data scientists use professional tools. In Python, the `scikit-learn` library provides `KNeighborsRegressor`:

        ```python
        from sklearn.neighbors import KNeighborsRegressor
        from sklearn.preprocessing import StandardScaler

        # Create and fit the model
        model = KNeighborsRegressor(n_neighbors=5)
        model.fit(X_train_scaled, y_train)

        # Make predictions
        predictions = model.predict(X_test_scaled)
        ```
        """
    },
    "scene_15_the_power_of_pipelines": {
        "title": "Scene 15: The Power of Pipelines",
        "markdown": """
        ## Scene 15: The Power of Pipelines
        *Professional detective workflows...*

        Real machine learning involves many preprocessing steps. Scikit-learn's **Pipeline** lets us chain operations together:

        ```python
        from sklearn.pipeline import make_pipeline

        pipeline = make_pipeline(
            StandardScaler(),           # First: scale the data
            KNeighborsRegressor(n_neighbors=5)  # Then: apply KNN
        )

        pipeline.fit(X_train, y_train)
        predictions = pipeline.predict(X_test)
        ```

        This ensures the same preprocessing is applied consistently to training and test data.
        """,
        "app_code": process_code_block("""
            # Streamlit app: sklearn Pipeline + cross-validation over K with RMSE curve
            st.write("Explore how different K values impact the model's performance using cross-validation.")

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
        """, "Scene 15: The Power of Pipelines")
    },
    "quiz_break_pipeline_benefits": {
        "title": "Quiz Break: Pipeline Benefits",
        "markdown": """
        ## Quiz Break: Pipeline Benefits

        **What's the main advantage of using a Pipeline?**

        A) It runs faster than separate steps
        B) It ensures consistent preprocessing across training and test data
        C) It automatically chooses the best K value

        ✅ **Answer: B** — Pipelines prevent the common mistake of preprocessing training and test data differently.
        """
    },
    "scene_16_beyond_single_predictions": {
        "title": "Scene 16: Beyond Single Predictions",
        "markdown": """
        ## Scene 16: Beyond Single Predictions
        *Scaling up our detective work...*

        Our method doesn't stop at predicting one wine. We can predict the quality of all the mystery vintages (1981-1991) by applying the same process to each year.

        The algorithm finds the nearest neighbors for each test vintage and makes individual predictions based on their specific weather patterns.
        """
    },
    "scene_17_regression_vs_classification": {
        "title": "Scene 17: Regression vs. Classification",
        "markdown": """
        ## Scene 17: Regression vs. Classification
        *Two types of detective problems...*

        **Regression problems** (like ours): Predict numerical values
        - Wine prices
        - House values
        - Temperature

        **Classification problems**: Predict categories
        - Email spam or not spam
        - Image contains cat or dog
        - Wine quality: excellent/good/poor

        K-Nearest Neighbors works for both! For classification, instead of averaging prices, we'd take a majority vote among the K neighbors.
        """,
        "app_code": process_code_block("""
            # Streamlit app: switch between regression and classification with the same KNN idea
            st.write("Explore KNN's use for both predicting continuous values (regression) and categories (classification).")

            rng = np.random.default_rng(123)
            N = 200
            temp = rng.normal(16.2, 0.6, N)
            rain = rng.normal(670, 80, N)
            price = 4.5 + 2.4*(temp-16) - 0.0035*(rain-670) + rng.normal(0, 1.2, N)
            quality_label = np.where(price > np.median(price), "high", "low")

            X = np.c_[temp, rain]
            mode = st.radio("Task", ["Regression (predict price)","Classification (high/low quality)"])
            K = st.slider("K", 1, 25, 7)

            if "Regression" in mode:
                model = make_pipeline(StandardScaler(), KNeighborsRegressor(n_neighbors=K))
                model.fit(X, price)
                st.write("Try adjusting a hypothetical vintage:")
                t = st.slider("Temp (°C)", 15.0, 17.5, 16.3, 0.1)
                r = st.slider("Rain (mm)", 520.0, 780.0, 720.0, 5.0)
                pred = model.predict([[t, r]])[0]
                st.metric("Predicted price", f"${pred:,.2f}")
            else:
                model = make_pipeline(StandardScaler(), KNeighborsClassifier(n_neighbors=K))
                model.fit(X, quality_label)
                t = st.slider("Temp (°C)", 15.0, 17.5, 16.3, 0.1)
                r = st.slider("Rain (mm)", 520.0, 780.0, 720.0, 5.0)
                pred = model.predict([[t, r]])[0]
                proba = model.predict_proba([[t, r]])[0]
                st.metric("Predicted class", pred)
                st.write(f"Probabilities → low: {proba[0]:.2f}, high: {proba[1]:.2f}")
        """, "Scene 17: Regression vs. Classification")
    },
    "quiz_break_problem_types": {
        "title": "Quiz Break: Problem Types",
        "markdown": """
        ## Quiz Break: Problem Types

        **Which of these would be a classification problem?**

        A) Predicting tomorrow's stock price
        B) Predicting whether a wine is red or white based on chemical properties
        C) Predicting a house's sale price

        ✅ **Answer: B** — Red vs. white are categories, not numerical values.
        """
    },
    "scene_18_the_algorithms_strengths": {
        "title": "Scene 18: The Algorithm's Strengths",
        "markdown": """
        ## Scene 18: The Algorithm's Strengths
        *Why K-Nearest Neighbors is powerful...*

        **Advantages:**
        - **Intuitive**: Easy to understand and explain
        - **No assumptions**: Works with any data pattern
        - **Flexible**: Handles both regression and classification
        - **Non-parametric**: Doesn't assume a specific mathematical form

        It's like having a wise friend who remembers every similar situation and gives advice based on how those turned out.
        """
    },
    "scene_19_the_algorithms_limitations": {
        "title": "Scene 19: The Algorithm's Limitations",
        "markdown": """
        ## Scene 19: The Algorithm's Limitations
        *Every detective method has weaknesses...*

        **Challenges:**
        - **Computational cost**: Must calculate distances to all training examples
        - **Sensitive to irrelevant features**: All features influence distance equally
        - **Curse of dimensionality**: Performance degrades with too many features
        - **Requires good K choice**: Too small = noisy, too large = overly smooth
        """
    },
    "scene_20_the_plot_twist_revealed": {
        "title": "Scene 20: The Plot Twist Revealed!",
        "markdown": """
        ## Scene 20: The Plot Twist Revealed!
        *History's surprise ending...*

        Here's the shocking revelation: **Ashenfelter's prediction was completely wrong!**

        The 1986 Bordeaux vintage, which K-Nearest Neighbors predicted would be mediocre (scoring around $13), actually became one of the greatest vintages ever produced, with a quality score of **95 out of 100**!

        Sometimes the most similar historical cases don't predict the future perfectly.
        """
    },
    "scene_21_why_predictions_fail": {
        "title": "Scene 21: Why Predictions Fail",
        "markdown": """
        ## Scene 21: Why Predictions Fail
        *The lesson in our detective story...*

        **Why was the prediction so far off?**

        1.  **Limited features**: Weather alone doesn't capture everything about wine quality
        2.  **Historical patterns**: The past doesn't always predict the future perfectly
        3.  **Small dataset**: With only ~30 training examples, patterns might be misleading
        4.  **Complex interactions**: Wine quality depends on subtle combinations weather factors

        This teaches us that machine learning models are tools, not crystal balls.
        """
    },
    "quiz_break_learning_from_failure": {
        "title": "Quiz Break: Learning from Failure",
        "markdown": """
        ## Quiz Break: Learning from Failure

        **What's the most important lesson from the 1986 prediction failure?**

        A) K-Nearest Neighbors is a bad algorithm
        B) We should always use more neighbors
        C) Models are imperfect tools that require human judgment

        ✅ **Answer: C** — Even sophisticated algorithms can fail. The key is understanding their limitations and using them wisely.
        """
    },
    "scene_22_choosing_k_in_practice": {
        "title": "Scene 22: Choosing K in Practice",
        "markdown": """
        ## Scene 22: Choosing K in Practice
        *The art of parameter selection...*

        How do we choose the best K value? Common strategies:

        **Small K (1-3):**
        - More sensitive to individual examples
        - Can capture fine details but may overfit to noise

        **Large K (10+):**
        - Smoother predictions
        - Less sensitive to outliers but may miss important patterns

        **The sweet spot:** Often found through **cross-validation** — testing different K values on held-out data.
        """
    },
    "scene_23_feature_engineering_detective_work": {
        "title": "Scene 23: Feature Engineering Detective Work",
        "markdown": """
        ## Scene 23: Feature Engineering Detective Work
        *Making better predictions...*

        To improve wine predictions, we might collect additional features:
        - **Soil composition**
        - **Harvest timing**
        - **Grape variety percentages**
        - **Vineyard altitude**
        - **Historical vintage reputation**

        More relevant features often lead to better neighbors and more accurate predictions.
        """
    },
    "scene_24_distance_metrics_beyond_euclidean": {
        "title": "Scene 24: Distance Metrics Beyond Euclidean",
        "markdown": """
        ## Scene 24: Distance Metrics Beyond Euclidean
        *Different ways to measure similarity...*

        **Euclidean distance** isn't the only option:

        - **Manhattan distance**: Sum of absolute differences
        - **Cosine similarity**: Angle between feature vectors
        - **Weighted distances**: Give some features more importance

        The choice depends on your data and problem context.
        """,
        "app_code": process_code_block("""
            # Streamlit app: compare distance metrics (Euclidean vs Manhattan vs Cosine) on prediction
            st.write("Compare how different distance metrics affect the prediction by finding different sets of 'nearest' neighbors.")

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

            st.metric("Prediction for 1986", f"${pred:,.2f}")
            st.dataframe(train.iloc[idx].assign(distance=d[idx])[["year","temp","rain","price","distance"]], use_container_width=True)
        """, "Scene 24: Distance Metrics Beyond Euclidean")
    },
    "scene_26_the_scalability_challenge": {
        "title": "Scene 26: The Scalability Challenge",
        "markdown": """
        ## Scene 26: The Scalability Challenge
        *When datasets grow large...*

        With millions of data points, calculating distances to every single example becomes expensive. Solutions include:

        **Approximate methods:**
        - Locality Sensitive Hashing
        - Random sampling of neighbors
        - Tree-based search structures

        **Preprocessing:**
        - Dimensionality reduction
        - Feature selection
        - Data clustering
        """
    },
    "quiz_break_scalability": {
        "title": "Quiz Break: Scalability",
        "markdown": """
        ## Quiz Break: Scalability

        **If you had 1 million training examples and needed to make predictions in real-time, what would be your biggest concern with basic K-Nearest Neighbors?**

        A) Memory usage
        B) Calculation time for distance to all examples
        C) Accuracy would decrease

        ✅ **Answer: B** — Computing distances to every single training example would be too slow for real-time applications.
        """
    },
    "scene_27_cross_validation_for_model_selection": {
        "title": "Scene 27: Cross-Validation for Model Selection",
        "markdown": """
        ## Scene 27: Cross-Validation for Model Selection
        *Testing our detective method properly...*

        To find the best K value:

        1.  **Split training data** into smaller training and validation sets
        2.  **Try different K values** (1, 3, 5, 7, 9, ...)
        3.  **Measure prediction accuracy** on validation set for each K
        4.  **Choose the K** with the best validation performance
        5.  **Final test** on completely unseen test data

        This prevents overfitting to any particular dataset split.
        """,
        "app_code": process_code_block("""
            # Streamlit app: grid search over K with holdout set + learning curve style plot
            st.write("Use cross-validation to find the optimal 'K' value for a synthetic dataset.")

            rng = np.random.default_rng(7)
            N = st.slider("Data size", 60, 600, 180, step=20)
            temp = rng.normal(16.2, 0.6, N)
            rain = rng.normal(670, 90, N)
            price = 4.5 + 2.4*(temp-16) - 0.0035*(rain-670) + rng.normal(0, 1.3, N)

            X = np.c_[temp, rain]
            y = price
            Xtr, Xte, ytr, yte = train_test_split(X, y, test_size=0.25, random_state=0)

            k_values = range(1, 32, 2)
            rmse = []
            scaler = StandardScaler().fit(Xtr)
            Xtrz, Xtez = scaler.transform(Xtr), scaler.transform(Xte)

            for k in k_values:
                knn = KNeighborsRegressor(n_neighbors=k)
                knn.fit(Xtrz, ytr)
                yp = knn.predict(Xtez)
                rmse.append(np.sqrt(mean_squared_error(yte, yp)))

            df = pd.DataFrame({"K":list(k_values),"RMSE":rmse})
            fig = px.line(df, x="K", y="RMSE", markers=True, title="Holdout RMSE vs K")
            st.plotly_chart(fig, use_container_width=True)
            st.success(f"Best K on holdout: {int(df.loc[df.RMSE.idxmin(),'K'])}")
        """, "Scene 27: Cross-Validation for Model Selection")
    },
    "scene_28_handling_different_data_types": {
        "title": "Scene 28: Handling Different Data Types",
        "markdown": """
        ## Scene 28: Handling Different Data Types
        *Not all features are numbers...*

        What if we have categorical features like:
        - **Wine region**: Bordeaux, Burgundy, Champagne
        - **Grape variety**: Cabernet, Merlot, Pinot Noir
        - **Season**: Spring, Summer, Fall, Winter

        **Solutions:**
        - **One-hot encoding**: Convert categories to binary features
        - **Ordinal encoding**: For ordered categories (Low, Medium, High)
        - **Specialized distance metrics**: That handle mixed data types
        """
    },
    "scene_29_the_curse_of_dimensionality": {
        "title": "Scene 29: The Curse of Dimensionality",
        "markdown": """
        ## Scene 29: The Curse of Dimensionality
        *When more features hurt performance...*

        Paradoxically, adding too many features can make K-Nearest Neighbors **worse**:

        - In high dimensions, all points become roughly equidistant
        - \"Nearest\" neighbors aren't really that near anymore
        - Random chance starts dominating meaningful patterns

        **The lesson:** More data isn't always better — focus on the most relevant features.
        """
    },
    "scene_30_weighted_k_nearest_neighbors": {
        "title": "Scene 30: Weighted K-Nearest Neighbors",
        "markdown": """
        ## Scene 30: Weighted K-Nearest Neighbors
        *Giving closer neighbors more influence...*

        Instead of treating all K neighbors equally, we can weight them by distance:

        - **Closer neighbors** get higher weights
        - **Farther neighbors** get lower weights
        - **Formula**: weight = 1/distance (or similar)

        This often provides more nuanced predictions than simple averaging.
        """,
        "app_code": process_code_block("""
            # Streamlit app: visualize neighbor contributions under distance weighting
            st.write("See how assigning more influence to closer neighbors changes a prediction.")

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
            st.metric("Weighted KNN Prediction", f"${pred:,.2f}")

            fig = px.bar(nn, x="year", y="contribution", hover_data=["price","weight"], title="Neighbor contributions to prediction")
            st.plotly_chart(fig, use_container_width=True)
            st.dataframe(nn[["year","price","weight","contribution"]], use_container_width=True)
        """, "Scene 30: Weighted K-Nearest Neighbors")
    },
    "quiz_break_weighting_intuition": {
        "title": "Quiz Break: Weighting Intuition",
        "markdown": """
        ## Quiz Break: Weighting Intuition

        **Why might weighted K-Nearest Neighbors give better predictions than regular KNN?**

        A) It uses fewer neighbors
        B) It gives more influence to the most similar cases
        C) It's faster to compute

        ✅ **Answer: B** — Intuitively, cases that are more similar should have more influence on our prediction.
        """
    },
    "scene_31_local_vs_global_patterns": {
        "title": "Scene 31: Local vs. Global Patterns",
        "markdown": """
        ## Scene 31: Local vs. Global Patterns
        *Understanding KNN's perspective...*

        K-Nearest Neighbors is a **local** method:
        - It only cares about nearby examples for each prediction
        - Different regions of feature space can have different patterns
        - No single global model is fit to all data

        This contrasts with **global** methods like linear regression that assume one pattern fits everywhere.
        """
    },
    "scene_32_handling_missing_data": {
        "title": "Scene 32: Handling Missing Data",
        "markdown": """
        ## Scene 32: Handling Missing Data
        *Real-world data is messy...*

        What if some vintages are missing temperature or rainfall data?

        **Options:**
        1.  **Remove examples** with missing features (loses information)
        2.  **Impute missing values** using KNN itself!
        3.  **Use distance metrics** that handle missing values specially
        4.  **Separate models** for different feature combinations
        """,
        "app_code": process_code_block("""
            # Streamlit app: simulate missing values and impute with KNNImputer
            st.write("Observe how KNNImputer can fill in missing data points based on the values of nearest neighbors.")

            rng = np.random.default_rng(8)
            N = 120
            temp = rng.normal(16.2, 0.6, N)
            rain = rng.normal(670, 80, N)
            price = 4.5 + 2.4*(temp-16) - 0.0035*(rain-670) + rng.normal(0, 1.2, N)

            df = pd.DataFrame({"temp":temp, "rain":rain, "price":price})
            frac = st.slider("Fraction of missing entries", 0.0, 0.5, 0.15, 0.05)
            mask = rng.random(df.size).reshape(df.shape) < frac
            df_missing = df.copy()
            df_missing.values[mask] = np.nan

            k = st.slider("Imputer K", 1, 15, 5)
            imputer = KNNImputer(n_neighbors=k)
            df_imputed = pd.DataFrame(imputer.fit_transform(df_missing), columns=df.columns)

            st.write("Original (head)")
            st.dataframe(df.head(10), use_container_width=True)
            st.write("With missing values (head)")
            st.dataframe(df_missing.head(10), use_container_width=True)
            st.write("Imputed (head)")
            st.dataframe(df_imputed.head(10), use_container_width=True)
        """, "Scene 32: Handling Missing Data")
    },
    "scene_33_the_bias_variance_tradeoff": {
        "title": "Scene 33: The Bias-Variance Tradeoff",
        "markdown": """
        ## Scene 33: The Bias-Variance Tradeoff
        *A fundamental machine learning concept...*

        **Small K (like K=1):**
        - **Low bias**: Can capture complex, local patterns
        - **High variance**: Very sensitive to individual training examples

        **Large K:**
        - **Higher bias**: Smoother, potentially over-simplified predictions
        - **Lower variance**: More stable across different training sets

        The optimal K balances these competing concerns.
        """,
        "app_code": process_code_block("""
            # Streamlit app: bias-variance with 1D function (synthetic), K slider
            st.write("Adjust 'K' to observe the bias-variance trade-off in a simple 1D regression problem.")

            rng = np.random.default_rng(0)
            n = st.slider("Training points", 20, 200, 80, step=10)
            noise = st.slider("Noise σ", 0.0, 1.0, 0.3, 0.05)
            K = st.slider("K", 1, 51, 9, step=2)

            # true function
            X = np.linspace(-3, 3, 200).reshape(-1,1)
            y_true = np.sin(X).ravel()

            # sample training data
            Xtr = rng.uniform(-3, 3, size=(n,1))
            ytr = np.sin(Xtr).ravel() + rng.normal(0, noise, size=n)

            pipe = make_pipeline(StandardScaler(with_mean=True, with_std=True), KNeighborsRegressor(n_neighbors=K))
            pipe.fit(Xtr, ytr)
            yp = pipe.predict(X)

            df_plot = pd.DataFrame({"x":X.ravel(),"y_true":y_true,"y_pred":yp})
            fig = px.line(df_plot, x="x", y=["y_true","y_pred"], title="Effect of K on fit")
            st.plotly_chart(fig, use_container_width=True)

            st.caption("Small K → flexible but noisy. Large K → smoother but biased.")
        """, "Scene 33: The Bias-Variance Tradeoff")
    },
 "scene_34_ensemble_methods_with_knn": {
        "title": "Scene 34: Ensemble Methods with KNN",
        "markdown": """
        ## Scene 34: Ensemble Methods with KNN
        *Combining multiple detectives...*

        We can improve predictions by combining multiple K-Nearest Neighbors models:

        - **Different K values**: Average predictions from K=3, K=5, K=7
        - **Different distance metrics**: Combine Euclidean and Manhattan distances
        - **Different feature subsets**: Some models use weather, others use soil data
        - **Bootstrap aggregating**: Train on different random samples
        """,
        "app_code": process_code_block("""
            # Streamlit app: demonstrate ensemble voting vs. single KNN prediction
            st.write("Compare individual KNN models with an ensemble approach on a classification task.")

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

            # Create subplots for individual models
            fig_individual = go.Figure()
            fig_individual.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k1, showscale=False, opacity=0.3, colorscale='Viridis', name='K=1 Prediction', hoverinfo='none'))
            fig_individual.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k5, showscale=False, opacity=0.3, colorscale='Plasma', name='K=5 Prediction', hoverinfo='none'))
            fig_individual.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_k11, showscale=False, opacity=0.3, colorscale='Cividis', name='K=11 Prediction', hoverinfo='none'))
            fig_individual.add_trace(go.Scatter(x=X[:, 0], y=X[:, 1], mode='markers', marker=dict(color=y, colorscale='Viridis', symbol='circle', size=8), name='Data Points'))
            fig_individual.update_layout(title="Decision Boundaries for Individual KNN Models")

            st.plotly_chart(fig_individual, use_container_width=True)

            # Ensemble plot
            fig_ensemble = go.Figure()
            fig_ensemble.add_trace(go.Contour(x=np.arange(x_min, x_max, 0.05), y=np.arange(y_min, y_max, 0.05), z=Z_ensemble, showscale=False, opacity=0.3, colorscale='Viridis', name='Ensemble Prediction', hoverinfo='none'))
            fig_ensemble.add_trace(go.Scatter(x=X[:, 0], y=X[:, 1], mode='markers', marker=dict(color=y, colorscale='Viridis', symbol='circle', size=8), name='Data Points'))
            fig_ensemble.update_layout(title="Ensemble (Voting) Prediction")

            st.plotly_chart(fig_ensemble, use_container_width=True)

            st.write("Notice how the ensemble prediction smooths out the noisy, irregular boundaries of the K=1 model and the overly simplified boundaries of the K=11 model, creating a more robust and accurate result.")
        """, "Scene 34: Ensemble Methods with KNN")
    },
    "quiz_break_ensemble_benefits": {
        "title": "Quiz Break: Ensemble Benefits",
        "markdown": """
        ## Quiz Break: Ensemble Benefits

        **What's the main benefit of combining multiple KNN models?**

        A) Faster prediction time
        B) Using less memory
        C) More robust and accurate predictions

        ✅ **Answer: C** — Multiple models can compensate for each other's individual mistakes, leading to better overall performance.
        """
    },
    "scene_35_modern_variations_and_improvements": {
        "title": "Scene 35: Modern Variations and Improvements",
        "markdown": """
        ## Scene 35: Modern Variations and Improvements
        *The evolution continues...*

        **Recent advances in nearest neighbor methods:**

        - **Approximate nearest neighbors**: Trade slight accuracy for massive speed improvements
        - **Learned distance metrics**: Use neural networks to learn better distance functions
        - **Adaptive K selection**: Choose different K values for different regions of space
        - **GPU acceleration**: Parallel distance calculations for massive datasets
        """
    },
    "scene_36_when_to_use_k_nearest_neighbors": {
        "title": "Scene 36: When to Use K-Nearest Neighbors",
        "markdown": """
        ## Scene 36: When to Use K-Nearest Neighbors
        *Choosing the right tool for the job...*

        **KNN works well when:**
        - Data has local patterns (similar inputs → similar outputs)
        - You need explainable predictions (\"because wines X, Y, Z were similar...\")
        - Dataset size is manageable
        - Features are meaningful for similarity comparison

        **Consider alternatives when:**
        - You need very fast real-time predictions
        - Dataset has millions of examples
        - Features are high-dimensional and sparse
        - You need probabilistic predictions
        """,
        "app_code": process_code_block("""
            # Streamlit app: back-of-the-envelope scalability calculator for KNN
            st.write("Estimate the computational cost of brute-force KNN as dataset size and features grow.")

            N = st.slider("Training size (N)", 1_000, 5_000_000, 200_000, step=1_000)
            d = st.slider("Number of features (d)", 2, 2000, 32)
            q = st.slider("Queries per second needed", 1, 10000, 100)

            # simplistic FLOP estimate: ~ (2d) per distance * N distances * q
            flops = 2 * d * N * q
            st.metric("Estimated floating-point ops/sec for brute-force search", f"{flops:,.0f}")

            #**Rules of thumb:**
            #- Use approximate nearest neighbors or pre-indexing when `N` is large.
            # - Lower `d` via feature selection or dimensionality reduction.
        """, "Scene 36: When to Use K-Nearest Neighbors")
    },
    "scene_37_interpreting_knn_predictions": {
        "title": "Scene 37: Interpreting KNN Predictions",
        "markdown": """
        ## Scene 37: Interpreting KNN Predictions
        *Making the black box transparent...*

        One of KNN's biggest advantages is **interpretability**:

        \"We predict this 1986 wine will cost $13.20 because it's most similar to 1974 ($8.50), 1958 ($16.00), 1969 ($14.50), 1968 ($15.50), and 1980 ($11.50).\"

        You can show stakeholders the exact historical examples that drove each prediction.
        """,
        "app_code": process_code_block("""
            # Streamlit app: explanation card listing the neighbors + textual rationale
            st.write("See which specific historical wines are used to make the prediction for the target year.")

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
            target_year = st.selectbox("Explain prediction for:", options=[1986], index=0)
            K = st.slider("K", 1, 7, 5)

            x = df.query("year==@target_year")[["temp","rain"]].to_numpy()
            X = train[["temp","rain"]].to_numpy()
            y = train["price"].to_numpy()

            # standardize
            mu, sd = X.mean(axis=0), X.std(axis=0, ddof=0)
            Xz = (X - mu)/sd
            xz = (x - mu)/sd

            d = np.sqrt(((Xz - xz)**2).sum(axis=1))
            idx = np.argsort(d)[:K]
            neighbors = train.iloc[idx].assign(distance=d[idx])

            pred = neighbors["price"].mean()
            st.metric(f"Prediction for {target_year}", f"${pred:,.2f}")

            st.write("**Rationale:** we averaged prices from the most similar weather years below.")
            st.dataframe(neighbors[["year","temp","rain","price","distance"]], use_container_width=True)
        """, "Scene 37: Interpreting KNN Predictions")
    },
    "scene_38_preprocessing_scaling_effect": {
        "title": "Scene 38: Why Scaling Matters (Practical Tips)",
        "markdown": """
        ## Scene 38: Practical Implementation Tips
        *Lessons from the field...*

        **Best practices for successful KNN projects:**

        1.  **Always standardize** numerical features
        2.  **Start with small K** and increase gradually
        3.  **Use cross-validation** to choose K
        4.  **Consider feature selection** to remove noise
        5.  **Monitor computational performance** as data grows
        6.  **Validate on completely unseen data**
        """,
        "app_code": process_code_block("""
            # Streamlit app: scaling effect demo (how unscaled rainfall dominates)
            st.write("This app illustrates why scaling features is crucial for accurate distance calculations in KNN.")

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

            t = df.query("year==1986")[["temp","rain"]].iloc[0].to_numpy()
            A = df.dropna(subset=["price"])[["temp","rain"]].to_numpy()

            # distances with/without scaling
            mu, sd = A.mean(axis=0), A.std(axis=0, ddof=0)
            A_z = (A - mu)/sd
            t_z = (t - mu)/sd

            d_raw = np.sqrt(((A - t)**2).sum(axis=1))
            d_z = np.sqrt(((A_z - t_z)**2).sum(axis=1))

            view = df.dropna(subset=["price"])[["year","price"]].copy()
            view["dist_unscaled"] = d_raw
            view["dist_scaled"] = d_z
            view = view.sort_values("dist_unscaled")

            fig = px.scatter(view, x="dist_unscaled", y="dist_scaled", text="year",
                             title="Distances to 1986: unscaled vs scaled")
            fig.update_traces(textposition="top center")
            st.plotly_chart(fig, use_container_width=True)
            st.dataframe(view, use_container_width=True)
        """, "Scene 38: Why Scaling Matters")
    },
    "scene_39_the_detectives_toolkit_summary": {
        "title": "Scene 39: The Detective's Toolkit Summary",
        "markdown": """
        ## Scene 39: The Detective's Toolkit Summary
        *Wrapping up our investigation...*

        **K-Nearest Neighbors Algorithm:**
        - **Input**: Training data with features and labels, new data to predict, K value
        - **Process**: Find K most similar training examples, average their labels
        - **Output**: Prediction based on local neighborhood
        - **Strength**: Intuitive, flexible, interpretable
        - **Weakness**: Can be slow, sensitive to irrelevant features
        """
    },
    "final_quiz_putting_it_all_together": {
        "title": "Final Quiz: Putting It All Together",
        "markdown": """
        ## Final Quiz: Putting It All Together

        **You're tasked with predicting house prices using KNN. Your features are: square footage, number of bedrooms, age of house, and neighborhood name. What's your first preprocessing step?**

        A) Choose K=5
        B) Standardize numerical features and encode neighborhood categorically
        C) Remove the neighborhood feature

        ✅ **Answer: B** — Before any KNN analysis, you need to ensure all features contribute fairly to distance calculations through proper scaling and encoding.
        """
    },
    "scene_40_try_it_yourself_playground": {
        "title": "Scene 40: KNN Playground (Bring Your Own Vintages)",
        "markdown": """
        ## Scene 40: The Continuing Detective Story
        *Where do we go from here?*

        Our journey through K-Nearest Neighbors reveals a fundamental truth about machine learning: the best algorithms are often the simplest to understand.

        While deep neural networks grab headlines, KNN quietly powers recommendation systems, fraud detection, and countless other applications where interpretability and simplicity matter.

        The 1986 wine prediction may have failed, but it taught us something invaluable: **models are tools to augment human judgment, not replace it.**

        **Remember:** Every great data scientist starts with simple, interpretable methods like KNN before moving to more complex approaches. Master the fundamentals, and the advanced techniques become much clearer.

        ---

        ## Final Reflection

        **Key Takeaways from Our Wine Detective Story:**

        1.  **Similarity-based reasoning** is intuitive and powerful
        2.  **Feature scaling** is crucial for fair distance calculations
        3.  **Model validation** helps us choose optimal parameters
        4.  **Interpretability** makes models trustworthy and actionable
        5.  **Failures teach us** about model limitations and the need for human oversight

        The next time you get a recommendation from Netflix or Amazon, remember: there's probably a K-Nearest Neighbors algorithm quietly working behind the scenes, finding users similar to you and suggesting what they enjoyed.

        **Your turn:** What problem in your life could benefit from finding \"nearest neighbors\" in data? 🕵️‍♂️🍷
        """,
        "app_code": process_code_block("""
            # Streamlit app: full playground — upload your own vintage CSV, tune K, metric, scaling
            st.write("Upload your own data or use the demo data to experiment with KNN parameters and make predictions.")

            #**Expected columns**
            #- Training rows: `year, summer_temp_c, winter_rain_mm, price_usd`
            #- Test rows (price unknown): `year, summer_temp_c, winter_rain_mm` with `price_usd` blank/NaN.

            uploaded = st.file_uploader("Upload CSV (or use demo)", type=["csv"])
            if uploaded is not None:
                df = pd.read_csv(uploaded)
            else:
                df = pd.DataFrame([
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
                ], columns=["year","summer_temp_c","winter_rain_mm","price_usd"])

            st.dataframe(df, use_container_width=True, height=240)

            train = df.dropna(subset=["price_usd"]).copy()
            test = df[df["price_usd"].isna()].copy()

            if train.empty or test.empty:
                st.stop()

            k = st.slider("K", 1, 21, 5, step=2)
            weights = st.selectbox("Weights", ["uniform","distance"])
            scale = st.toggle("Standardize features", True)
            metric = st.selectbox("Metric", ["minkowski (euclidean)","manhattan"], index=0)

            Xtr = train[["summer_temp_c","winter_rain_mm"]].to_numpy()
            ytr = train["price_usd"].to_numpy()
            Xte = test[["summer_temp_c","winter_rain_mm"]].to_numpy()

            if scale:
                scaler = StandardScaler().fit(Xtr)
                Xtr = scaler.transform(Xtr)
                Xte = scaler.transform(Xte)

            p = 2 if "minkowski" in metric else 1
            model = KNeighborsRegressor(n_neighbors=k, weights=weights, p=p)
            model.fit(Xtr, ytr)
            preds = model.predict(Xte)

            out = test.copy()
            out["predicted_price_usd"] = preds
            st.subheader("Predictions")
            st.dataframe(out[["year","summer_temp_c","winter_rain_mm","predicted_price_usd"]], use_container_width=True)
        """, "Scene 40: KNN Playground (Bring Your Own Vintages)")
    },
}

# Define the order of scenes for the sidebar
ordered_scene_ids = [
    "intro",
    "scene_04_visualization_reveals_all",
    "scene_05_the_1986_mystery",
    "scene_06_formalizing_our_detective_work",
    "scene_07_10_11_knn_distance_and_neighbors",
    "quiz_break_algorithm_understanding",
    "scene_13_implementing_from_scratch",
    "scene_14_the_professional_detective_tools",
    "scene_15_the_power_of_pipelines",
    "quiz_break_pipeline_benefits",
    "scene_16_beyond_single_predictions",
    "scene_17_regression_vs_classification",
    "quiz_break_problem_types",
    "scene_18_the_algorithms_strengths",
    "scene_19_the_algorithms_limitations",
    "scene_20_the_plot_twist_revealed",
    "scene_21_why_predictions_fail",
    "quiz_break_learning_from_failure",
    "scene_22_choosing_k_in_practice",
    "scene_23_feature_engineering_detective_work",
    "scene_24_distance_metrics_beyond_euclidean",
    "scene_26_the_scalability_challenge",
    "quiz_break_scalability",
    "scene_27_cross_validation_for_model_selection",
    "scene_28_handling_different_data_types",
    "scene_29_the_curse_of_dimensionality",
    "scene_30_weighted_k_nearest_neighbors",
    "quiz_break_weighting_intuition",
    "scene_31_local_vs_global_patterns",
    "scene_32_handling_missing_data",
    "scene_33_the_bias_variance_tradeoff",
    "scene_34_ensemble_methods_with_knn",
    "quiz_break_ensemble_benefits",
    "scene_35_modern_variations_and_improvements",
    "scene_36_when_to_use_k_nearest_neighbors",
    "scene_37_interpreting_knn_predictions",
    "scene_38_preprocessing_scaling_effect",
    "scene_39_the_detectives_toolkit_summary",
    "final_quiz_putting_it_all_together",
    "scene_40_try_it_yourself_playground",
]

# --- Sidebar Navigation ---
st.sidebar.title("Scenes")
# Create a dictionary for sidebar radio options where display title is key and scene_id is value
sidebar_options = {scenes[sid]["title"]: sid for sid in ordered_scene_ids}

selected_title = st.sidebar.radio("Navigate through the story:", list(sidebar_options.keys()))
selected_scene_id = sidebar_options[selected_title]

# --- Main Content Display ---
def show_scene(scene_id):
    scene_data = scenes[scene_id]
    st.markdown(scene_data["markdown"])
    
    if "app_code" in scene_data:
        st.markdown("---")
        try:
            # Execute the processed code within the current Streamlit context
            # This allows the embedded app to run directly
            exec(scene_data["app_code"])
        except Exception as e:
            st.error(f"An error occurred while running the app for '{scene_data['title']}': {e}")
            st.exception(e) # Display full traceback for debugging
        st.markdown("---")
        st.markdown("#### Code for this interactive app:")
        st.code(scene_data["app_code"], language="python")

# Display the selected scene
show_scene(selected_scene_id)