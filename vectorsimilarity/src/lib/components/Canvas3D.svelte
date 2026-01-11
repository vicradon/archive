<script>
    import { onMount } from 'svelte';
    
    let { vectors = [], selectedVector = null, onselect } = $props();

    /**
     * @typedef {HTMLDivElement & { on: (event: string, callback: (data: any) => void) => void }} PlotlyHTMLElement
     */

    /** @type {PlotlyHTMLElement | HTMLDivElement} */
    let plotDiv;

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
             const interval = setInterval(() => {
                if (window.Plotly) {
                    clearInterval(interval);
                    draw();
                    setupEvents();
                }
            }, 100);
        }
    });

    function setupEvents() {
        plotDiv.on('plotly_click', (data) => {
             if (data.points && data.points.length > 0) {
                 const pointIdx = data.points[0].pointIndex;
                 const curveIdx = data.points[0].curveNumber;
                 // If we have a dedicated "Heads" trace at index 0
                 if (curveIdx === 0 && vectors[pointIdx]) {
                     onselect(vectors[pointIdx]);
                 }
            }
        });
    }

    function draw() {
        if (!plotDiv || !window.Plotly) return;

        const isDark = document.documentElement.classList.contains('dark');
        const fgColor = isDark ? '#fafafa' : '#09090b';
        const gridColor = isDark ? '#27272a' : '#e4e4e7';

        // 1. Trace for click detection (Markers at tips)
        const headsTrace = {
            x: vectors.map(v => v.components[0]),
            y: vectors.map(v => v.components[1]),
            z: vectors.map(v => v.components[2]),
            mode: 'markers',
            type: 'scatter3d',
            marker: {
                size: 5,
                color: vectors.map(v => v.color),
                opacity: 0.5 
            },
            hovertext: vectors.map(v => v.name),
            name: 'Tips'
        };

        /** @type {any[]} */
        const data = [headsTrace];

        // 2. Add traces for each vector (Line + Cone)
        // Let's use lines for shafts.
        vectors.forEach(v => {
            data.push({
                type: 'scatter3d',
                mode: 'lines',
                x: [0, v.components[0]],
                y: [0, v.components[1]],
                z: [0, v.components[2]],
                line: {
                    color: v.color,
                    width: v === selectedVector ? 10 : 5
                },
                showlegend: false,
                hoverinfo: 'none'
            });
        });

        // 3. Cones for heads
        if (vectors.length > 0) {
             vectors.forEach(v => {
                data.push({
                    type: 'cone',
                    x: [v.components[0]],
                    y: [v.components[1]],
                    z: [v.components[2]],
                    u: [v.components[0]], 
                    v: [v.components[1]], 
                    w: [v.components[2]],
                    sizemode: 'absolute',
                    sizeref: 0.5,
                    anchor: 'tip',
                    colorscale: [[0, v.color], [1, v.color]],
                    showscale: false,
                    hoverinfo: 'none'
                });
             });
        }

        const layout = {
            margin: { t: 0, r: 0, b: 0, l: 0 },
            paper_bgcolor: 'rgba(0,0,0,0)',
            scene: {
                xaxis: { title: 'X', gridcolor: gridColor, zerolinecolor: fgColor },
                yaxis: { title: 'Y', gridcolor: gridColor, zerolinecolor: fgColor },
                zaxis: { title: 'Z', gridcolor: gridColor, zerolinecolor: fgColor },
                aspectmode: 'cube'
            },
            showlegend: false
        };

        window.Plotly.react(plotDiv, data, layout, { displayModeBar: true, responsive: true });
    }

</script>

<div bind:this={plotDiv} style="width: 100%; height: 100%; min-height: 400px;"></div>
