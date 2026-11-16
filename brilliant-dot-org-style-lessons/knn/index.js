        let currentSection = 1;
        const totalSections = 6;
        let wineData = [];
        let testPoint = null;
        let k = 3;

        // Generate realistic wine data based on Ashenfelter's research
        function generateWineData() {
            const data = [];
            
            // High-quality vintage years (hot summers, moderate winter rain)
            for (let i = 0; i < 15; i++) {
                data.push({
                    summer: 16.5 + Math.random() * 2.5, // Hot summers
                    winter: 500 + Math.random() * 200,  // Moderate winter rain
                    price: 45 + Math.random() * 25,     // Expensive wines
                    year: 1961 + Math.floor(Math.random() * 50)
                });
            }
            
            // Good quality wines (decent summers, variable rain)
            for (let i = 0; i < 25; i++) {
                data.push({
                    summer: 14 + Math.random() * 4,
                    winter: 300 + Math.random() * 500,
                    price: 20 + Math.random() * 30,
                    year: 1950 + Math.floor(Math.random() * 60)
                });
            }
            
            // Lower quality wines (cool summers, extreme rain conditions)
            for (let i = 0; i < 15; i++) {
                data.push({
                    summer: 11 + Math.random() * 3,
                    winter: Math.random() < 0.5 ? 100 + Math.random() * 200 : 700 + Math.random() * 200,
                    price: 5 + Math.random() * 20,
                    year: 1940 + Math.floor(Math.random() * 70)
                });
            }
            
            return data;
        }

        function initializeLesson() {
            wineData = generateWineData();
            updateProgress();
            setupWinePlot();
            updateSectionIndicator();
            
            // Setup k slider
            document.getElementById('kSlider').addEventListener('input', function(e) {
                k = parseInt(e.target.value);
                document.getElementById('kValue').textContent = k;
                if (testPoint) {
                    highlightNearestNeighbors();
                    updatePrediction();
                }
            });

            // Initialize theme
            if (localStorage.getItem('theme') === 'dark') {
                document.documentElement.classList.add('dark');
            }
        }

        function setupWinePlot() {
            const plot = document.getElementById('winePlot') || document.getElementById('predictionPlot');
            if (!plot) return;
            
            plot.innerHTML = '<div class="axis-label y-axis">Summer Temperature (°C)</div>';
            
            wineData.forEach((wine, index) => {
                const point = document.createElement('div');
                point.className = 'wine-point';
                
                // Normalize coordinates to plot area
                const x = ((wine.winter - 50) / 850) * 100;
                const y = 100 - ((wine.summer - 10) / 10) * 100;
                
                point.style.left = `${Math.max(2, Math.min(98, x))}%`;
                point.style.top = `${Math.max(2, Math.min(98, y))}%`;
                
                // Color based on price (green = cheap, red = expensive)
                const priceNormalized = Math.min(wine.price / 80, 1);
                const hue = 120 * (1 - priceNormalized); // 120 = green, 0 = red
                point.style.backgroundColor = `hsl(${hue}, 70%, 50%)`;
                
                point.title = `Year: ${wine.year}\nSummer: ${wine.summer.toFixed(1)}°C\nWinter: ${wine.winter.toFixed(0)}mm\nPrice: ${wine.price.toFixed(1)}`;
                point.setAttribute('data-index', index);
                
                plot.appendChild(point);
            });
            
            // Add click listener for test points
            plot.addEventListener('click', function(e) {
                if (e.target === plot || e.target.classList.contains('axis-label')) {
                    const rect = plot.getBoundingClientRect();
                    const x = (e.clientX - rect.left) / rect.width;
                    const y = (e.clientY - rect.top) / rect.height;
                    
                    addTestPoint(x, y);
                }
            });
        }

        function addTestPoint(x, y) {
            const plot = document.getElementById('winePlot') || document.getElementById('predictionPlot');
            
            // Remove existing test point
            const existing = plot.querySelector('.test-point');
            if (existing) existing.remove();
            
            // Add new test point
            const point = document.createElement('div');
            point.className = 'wine-point test-point';
            point.style.left = `${x * 100}%`;
            point.style.top = `${y * 100}%`;
            
            testPoint = {
                x: x,
                y: y,
                summer: 10 + (1 - y) * 10,
                winter: 50 + x * 850
            };
            
            point.title = `Test Point\nSummer: ${testPoint.summer.toFixed(1)}°C\nWinter: ${testPoint.winter.toFixed(0)}mm`;
            
            plot.appendChild(point);
            
            setTimeout(() => {
                highlightNearestNeighbors();
                updatePrediction();
            }, 100);
        }

        function highlightNearestNeighbors() {
            if (!testPoint) return;
            
            const plot = document.getElementById('winePlot') || document.getElementById('predictionPlot');
            const points = plot.querySelectorAll('.wine-point:not(.test-point)');
            
            // Calculate distances (using scaled features)
            const distances = wineData.map((wine, index) => {
                // Scale features to prevent one from dominating
                const dx = (testPoint.summer - wine.summer) / 10; // temperature range ~10°C
                const dy = (testPoint.winter - wine.winter) / 850; // rainfall range ~850mm
                const distance = Math.sqrt(dx * dx + dy * dy);
                return { index, distance, price: wine.price, wine };
            });
            
            distances.sort((a, b) => a.distance - b.distance);
            
            // Clear previous highlights
            points.forEach(point => point.classList.remove('neighbor'));
            
            // Highlight k nearest neighbors
            for (let i = 0; i < k && i < distances.length; i++) {
                const pointElement = plot.querySelector(`[data-index="${distances[i].index}"]`);
                if (pointElement) {
                    pointElement.classList.add('neighbor');
                }
            }
            
            return distances.slice(0, k);
        }

        function updatePrediction() {
            if (!testPoint) return;
            
            const neighbors = highlightNearestNeighbors();
            const avgPrice = neighbors.reduce((sum, n) => sum + n.price, 0) / neighbors.length;
            
            const resultDiv = document.getElementById('predictionResult');
            if (resultDiv) {
                resultDiv.style.display = 'block';
                document.getElementById('predictedPrice').textContent = avgPrice.toFixed(1);
                document.getElementById('neighborsCount').textContent = k;
            }
        }

        function demonstrateAlgorithm() {
            const demo = document.getElementById('algorithmDemo');
            demo.innerHTML = '';
            
            if (!testPoint) {
                demo.innerHTML = '<div style="color: var(--geist-red-9); padding: 16px; background: var(--geist-background-2); border-radius: 8px; border: 2px solid var(--geist-red-9);">Please add a test point first by clicking on the plot above! 👆</div>';
                return;
            }
            
            const steps = [
                'Calculating Euclidean distances to all training points...',
                `Finding the ${k} nearest neighbors and highlighting them...`,
                'Averaging their wine prices to make the final prediction...'
            ];
            
            let stepIndex = 0;
            const stepElements = document.querySelectorAll('.step');
            
            function showNextStep() {
                if (stepIndex < steps.length) {
                    // Update active step
                    stepElements.forEach(el => el.classList.remove('active'));
                    if (stepElements[stepIndex]) {
                        stepElements[stepIndex].classList.add('active');
                    }
                    
                    const stepDiv = document.createElement('div');
                    stepDiv.style.cssText = `
                        margin: 12px 0; 
                        padding: 16px; 
                        background: var(--geist-background-1); 
                        border: 2px solid var(--geist-blue-9); 
                        border-radius: 8px; 
                        animation: slideIn 0.4s ease;
                        font-weight: 500;
                    `;
                    stepDiv.textContent = `Step ${stepIndex + 1}: ${steps[stepIndex]}`;
                    demo.appendChild(stepDiv);
                    
                    if (stepIndex === 1) {
                        highlightNearestNeighbors();
                    } else if (stepIndex === 2) {
                        updatePrediction();
                    }
                    
                    stepIndex++;
                    setTimeout(showNextStep, 2000);
                }
            }
            
            showNextStep();
        }

        function selectAnswer(element, isCorrect) {
            const options = element.parentElement.querySelectorAll('.quiz-option');
            options.forEach(opt => {
                opt.style.pointerEvents = 'none';
                if (opt === element) {
                    opt.classList.add(isCorrect ? 'correct' : 'incorrect');
                } else {
                    // Highlight correct answers
                    const text = opt.textContent.toLowerCase();
                    if (text.includes('regression') || 
                        text.includes('cluster in certain') ||
                        text.includes('features with different') || 
                        text.includes('chains preprocessing') ||
                        text.includes('k = 5-10')) {
                        opt.classList.add('correct');
                    }
                }
            });

            // Show feedback
            setTimeout(() => {
                if (isCorrect) {
                    element.innerHTML += ' ✅';
                } else {
                    element.innerHTML += ' ❌';
                }
            }, 500);
        }

        function nextSection() {
            if (currentSection < totalSections) {
                document.getElementById(`section${currentSection}`).classList.remove('active');
                currentSection++;
                document.getElementById(`section${currentSection}`).classList.add('active');
                updateProgress();
                updateSectionIndicator();
                
                // Setup plot for new sections
                if (currentSection === 2 || currentSection === 4) {
                    setTimeout(setupWinePlot, 100);
                }
                
                // Scroll to top of new section
                document.getElementById(`section${currentSection}`).scrollIntoView({ 
                    behavior: 'smooth', 
                    block: 'start' 
                });
            }
        }

        function previousSection() {
            if (currentSection > 1) {
                document.getElementById(`section${currentSection}`).classList.remove('active');
                currentSection--;
                document.getElementById(`section${currentSection}`).classList.add('active');
                updateProgress();
                updateSectionIndicator();
                
                // Setup plot for previous sections
                if (currentSection === 2 || currentSection === 4) {
                    setTimeout(setupWinePlot, 100);
                }
                
                // Scroll to top of new section
                document.getElementById(`section${currentSection}`).scrollIntoView({ 
                    behavior: 'smooth', 
                    block: 'start' 
                });
            }
        }

        function updateProgress() {
            const progress = (currentSection / totalSections) * 100;
            document.getElementById('progressBar').style.width = `${progress}%`;
        }

        function updateSectionIndicator() {
            document.getElementById('sectionIndicator').textContent = `Section ${currentSection} of ${totalSections}`;
            
            // Update button states
            const prevBtn = document.getElementById('prevBtn');
            const nextBtn = document.getElementById('nextBtn');
            
            prevBtn.disabled = currentSection === 1;
            
            if (currentSection === totalSections) {
                nextBtn.textContent = 'Complete! 🎉';
                nextBtn.onclick = () => {
                    alert('🎉 Congratulations! You\'ve completed the K-Nearest Neighbors lesson!\n\nYou now understand:\n✅ How K-NN works\n✅ When to use it\n✅ How to implement it\n✅ Key considerations (scaling, choosing k)\n\nKeep learning and exploring machine learning!');
                };
            } else {
                nextBtn.textContent = 'Next →';
                nextBtn.onclick = nextSection;
            }
        }

        function toggleTheme() {
            document.documentElement.classList.toggle('dark');
            const isDark = document.documentElement.classList.contains('dark');
            localStorage.setItem('theme', isDark ? 'dark' : 'light');
        }

        // Initialize the lesson when page loads
        window.addEventListener('DOMContentLoaded', initializeLesson);

        // Add some keyboard navigation
        document.addEventListener('keydown', function(e) {
            if (e.key === 'ArrowRight' && currentSection < totalSections) {
                nextSection();
            } else if (e.key === 'ArrowLeft' && currentSection > 1) {
                previousSection();
            }
        });