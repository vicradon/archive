<script>

    import { onMount } from 'svelte';
    
    let { vectors = [], selectedVector = null, onselect, onadd, onupdate } = $props();

    /**
     * @typedef {HTMLDivElement & { on: (event: string, callback: (data: any) => void) => void }} PlotlyHTMLElement
     */

    /** @type {PlotlyHTMLElement | HTMLDivElement} */
    let plotDiv;
    
    // Drag state
    let draggingVector = $state(null);
    let hoverVector = null;

    $effect(() => {
        if (plotDiv && window.Plotly) {
            draw();
        }
    });

    onMount(() => {
        if (window.Plotly) {
            draw();
            setupEvents();
        } else {
            // Poll for Plotly if script is async/defer (unlikely with our current app.html change but safe)
            const interval = setInterval(() => {
                if (window.Plotly) {
                    clearInterval(interval);
                    draw();
                    setupEvents();
                }
            }, 100);
        }
        
        // Window events for drag
        window.addEventListener('pointermove', handlePointerMove);
        window.addEventListener('pointerup', handlePointerUp);
        
        return () => {
             window.removeEventListener('pointermove', handlePointerMove);
             window.removeEventListener('pointerup', handlePointerUp);
        };
    });

    function setupEvents() {
         plotDiv.on('plotly_click', (data) => {
            if (data.points && data.points.length > 0) {
                const pointIdx = data.points[0].pointIndex;
                if(onselect && vectors[pointIdx]) {
                   onselect(vectors[pointIdx]);
               }
            }
        });
        
        // Track hover for drag target
        plotDiv.on('plotly_hover', (data) => {
            if (data.points && data.points.length > 0) {
                 const pointIdx = data.points[0].pointIndex;
                 if (vectors[pointIdx]) {
                     hoverVector = vectors[pointIdx];
                 }
            }
        });
        
        plotDiv.on('plotly_unhover', () => {
            hoverVector = null;
        });
    }
    
    function handlePointerDown(e) {
        if (hoverVector) {
            draggingVector = hoverVector;
            if (onselect) onselect(draggingVector);
            // Disable Plotly drag/pan temporarily? 
            // We can preventDefault, but we need to verify if Plotly captures it first.
            // Actually, we want to start our drag.
            // If dragging, we should update the UI.
        }
    }
    
    function handlePointerMove(e) {
        if (!draggingVector || !plotDiv) return;
        
        // Convert screen coordinates to data coordinates
        // Using getBoundingClientRect because local offsets are safest
        const rect = plotDiv.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        // Accces Plotly internal axis to convert back
        // @ts-ignore
        const xaxis = plotDiv._fullLayout.xaxis;
        // @ts-ignore
        const yaxis = plotDiv._fullLayout.yaxis;
        
        const dataX = xaxis.p2d(x);
        const dataY = yaxis.p2d(y);
        
        // Update vector components
        draggingVector.components = [dataX, dataY];
        
        // Notify parent
        if (onupdate) {
            onupdate(draggingVector);
        }
    }
    
    function handlePointerUp() {
        draggingVector = null;
    }

    function draw() {
        if (!plotDiv || !window.Plotly) return;

        const isDark = document.documentElement.classList.contains('dark');
        const fgColor = isDark ? '#fafafa' : '#09090b';
        const gridColor = isDark ? '#27272a' : '#e4e4e7';
        
        const traces = [];

        // 0. Halo for selected vector (behind everything)
        // Make it very obvious: larger marker + thick line
        if (selectedVector) {
             const haloTrace = {
                x: [0, selectedVector.components[0]], // Line from origin
                y: [0, selectedVector.components[1]],
                mode: 'lines+markers',
                type: 'scatter',
                marker: {
                    size: 25, // Large halo
                    color: selectedVector.color,
                    opacity: 0.3,
                    line: { width: 0 }
                },
                line: {
                    color: selectedVector.color,
                    width: 10,
                    opacity: 0.3
                },
                hoverinfo: 'skip'
            };
            traces.push(haloTrace);
        }

        // Trace for Vector Heads (markers)
        // This allows clicking to select.
        const headsTrace = {
            x: vectors.map(v => v.components[0]),
            y: vectors.map(v => v.components[1]),
            mode: 'markers',
            type: 'scatter',
            marker: {
                color: vectors.map(v => v === selectedVector ? '#18181b' : v.color), // highlight selected
                size: 20, // larger hit area
                opacity: 0, // invisible but clickable
                line: {
                    color: '#ffffff',
                    width: 1
                }
            },
            hovertext: vectors.map(v => v.name),
            name: 'Vectors'
        };
        traces.push(headsTrace);

        // Arrows (Lines)
        // We can use 'shapes' layout attribute for lines from 0,0 to x,y
        const shapes = vectors.map(v => ({
            type: 'line',
            x0: 0,
            y0: 0,
            x1: v.components[0],
            y1: v.components[1],
            line: {
                color: v.color,
                width: v === selectedVector ? 4 : 2
            }
        }));

        // Annotations for Arrow Heads and Labels
        // We use two sets or one set?
        // Let's use annotations for the arrows themselves as before.
        // But for "labels along the body", we might want separate annotations or just text on the arrow annotation (which is at the tip usually).
        // If we want text along the body, we need a separate annotation at midpoint.
        
        /** @type {any[]} */
        const annotations = [];
        
        vectors.forEach(v => {
            // 1. Arrow Annotation (Visual arrow)
            annotations.push({
                x: v.components[0],
                y: v.components[1],
                xref: 'x',
                yref: 'y',
                ax: 0,
                ay: 0,
                axref: 'x',
                ayref: 'y',
                showarrow: true,
                arrowhead: 2,
                arrowsize: 1,
                arrowwidth: v === selectedVector ? 3 : 2,
                arrowcolor: v.color,
                // No text here, just the arrow
            });

            // 2. Label Annotation (Text along body)
            // Midpoint
            const midX = v.components[0] / 2;
            const midY = v.components[1] / 2;
            // Angle for text rotation
            let angle = Math.atan2(v.components[1], v.components[0]) * 180 / Math.PI;
            // Keep text upright-ish
            if (angle > 90) angle -= 180;
            if (angle < -90) angle += 180;

            annotations.push({
                x: midX,
                y: midY,
                xref: 'x',
                yref: 'y',
                text: v.name,
                showarrow: false,
                font: { 
                    color: v.color, 
                    size: 12,
                    family: 'sans-serif',
                    weight: 'bold' // make it pop
                },
                textangle: angle,
                bgcolor: isDark ? 'rgba(0,0,0,0.5)' : 'rgba(255,255,255,0.8)', // readability
                borderpad: 2
            });
        });

        // Grid range
        // Find max magnitude to set range. 
        // We can let Plotly autoscale initially, but "square" ratio is important.
        const maxVal = vectors.reduce((max, v) => Math.max(max, Math.abs(v.components[0]), Math.abs(v.components[1])), 5) + 2;

        const layout = {
            autosize: true,
            margin: { t: 20, r: 20, b: 20, l: 20 },
            paper_bgcolor: 'rgba(0,0,0,0)', 
            plot_bgcolor: 'rgba(0,0,0,0)',
            xaxis: {
                zeroline: true,
                zerolinecolor: fgColor,
                gridcolor: gridColor,
                // let user zoom, but setting range initialises it nicely
                autorange: false,
                range: [-maxVal, maxVal] 
            },
            yaxis: {
                zeroline: true,
                zerolinecolor: fgColor,
                gridcolor: gridColor,
                scaleanchor: 'x', 
                scaleratio: 1,
                autorange: false,
                range: [-maxVal, maxVal]
            },
            showlegend: false,
            annotations: annotations,
            dragmode: draggingVector ? false : 'pan', // Disable pan while dragging vector
            hovermode: 'closest'
        };
        
        window.Plotly.react(plotDiv, traces, layout, { 
            displayModeBar: true, // Show bar so they can reset zoom etc
            responsive: true,
            scrollZoom: true // Enable scroll to zoom!
        });
    }

</script>

<div bind:this={plotDiv} style="width: 100%; height: 100%; min-height: 400px;" onpointerdown={handlePointerDown}></div>
