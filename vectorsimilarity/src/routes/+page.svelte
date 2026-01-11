<script>
    import { onMount } from 'svelte';
    import { VectorMath } from '$lib/vector-math.js';
    import Canvas2D from '$lib/components/Canvas2D.svelte';
    import SimilarityMatrix from '$lib/components/SimilarityMatrix.svelte';

    let vectors = $state([]);
    let selectedVector = $state(null);
    let vectorCounter = 0;
    const colors = ['#ef4444', '#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'];
    
    // Inputs
    let vectorInput = $state('');
    
    // Derived values for controls (not strictly derived as we bind them or update them)
    // We can use $effect to sync, or getters?
    // Start with simple sync functions.

    function addVector(components) {
        vectorCounter++;
        const newVector = {
            id: vectorCounter,
            name: `v${vectorCounter}`,
            components,
            color: colors[vectors.length % colors.length]
        };
        vectors = [...vectors, newVector];
    }

    function addFromInput() {
        const parsed = VectorMath.parse(vectorInput);
        if (!parsed || parsed.length < 2 || parsed.some(isNaN)) {
            alert('Invalid format. Use [x, y] or [[x], [y]]');
            return;
        }
        addVector([parsed[0], parsed[1]]);
        vectorInput = '';
    }

    function selectVector(v) {
        selectedVector = v;
    }

    function deleteVector(id) {
        vectors = vectors.filter(v => v.id !== id);
        if (selectedVector && selectedVector.id === id) {
            selectedVector = null;
        }
    }
    
    function deleteSelected() {
        if (selectedVector) {
            deleteVector(selectedVector.id);
        }
    }

    function duplicateVector() {
        if (!selectedVector) return;
        // Shift slightly so it's visible
        addVector([selectedVector.components[0] + 1, selectedVector.components[1] + 1]);
    }

    function clearAll() {
        vectors = [];
        selectedVector = null;
    }

    // Update helpers for sliders
    function updateVectorMagnitude(mag) {
        if (!selectedVector) return;
        const currentMag = VectorMath.magnitude(selectedVector.components);
        if (currentMag === 0 && mag > 0) {
            selectedVector.components = [mag, 0];
        } else if (currentMag > 0) {
            const ratio = mag / currentMag;
            selectedVector.components = selectedVector.components.map(c => c * ratio);
        }
        vectors = [...vectors]; // trigger update
    }

    function updateVectorAngle(deg) {
        if (!selectedVector) return;
        const mag = VectorMath.magnitude(selectedVector.components);
        const rad = deg * Math.PI / 180;
        selectedVector.components = [
            mag * Math.cos(rad),
            mag * Math.sin(rad)
        ];
        vectors = [...vectors]; // trigger update
    }

    // Getters for slider values
    function getMagnitude(v) {
        if (!v) return 0;
        return VectorMath.magnitude(v.components);
    }

    function getAngle(v) {
        if (!v) return 0;
        const angle = Math.atan2(v.components[1], v.components[0]) * 180 / Math.PI;
        return angle < 0 ? angle + 360 : angle;
    }

</script>

<div class="main-layout">
    <div class="canvas-section">
        <div class="card canvas-wrapper">
            <Canvas2D 
                {vectors} 
                {selectedVector} 
                onselect={selectVector} 
                onadd={addVector}
                onupdate={(v) => vectors = vectors}
            />
        </div>

        {#if selectedVector}
            <!-- Mobile Controls -->
             <div class="card mobile-vector-controls" id="mobileVectorControls">
                <div class="card-header" style="position: relative;">
                    <h2>Selected: {selectedVector.name}</h2>
                    <button class="close-controls" onclick={() => selectedVector = null}>×</button>
                </div>
                <!-- Logic is same as desktop, ideally share component but keeping inline for speed -->
                 <div class="card-body">
                    <!-- Components -->
                     <div class="component-inputs">
                        <div class="component-input">
                            <label for="mCompX">X</label>
                            <input type="number" id="mCompX" class="input" step="0.1" 
                                value={selectedVector.components[0]} 
                                oninput={(e) => {
                                    selectedVector.components[0] = parseFloat(e.target.value) || 0;
                                    vectors = [...vectors];
                                }}
                            >
                        </div>
                        <div class="component-input">
                            <label for="mCompY">Y</label>
                            <input type="number" id="mCompY" class="input" step="0.1" 
                                value={selectedVector.components[1]}
                                oninput={(e) => {
                                    selectedVector.components[1] = parseFloat(e.target.value) || 0;
                                    vectors = [...vectors];
                                }}
                            >
                        </div>
                     </div>
                     
                     <!-- Sliders -->
                     <!-- To simplify binding with derived vals, we use value={get...} and oninput calls update... -->
                     <div class="slider-control">
                        <div class="value-input-group">
                            <label>Magnitude</label>
                            <input type="number" class="input" step="0.1" min="0" 
                                value={getMagnitude(selectedVector).toFixed(1)}
                                oninput={(e) => updateVectorMagnitude(parseFloat(e.target.value)||0)}
                            >
                        </div>
                        <input type="range" min="0" max="20" step="0.1" 
                            value={Math.min(getMagnitude(selectedVector), 20)}
                            oninput={(e) => updateVectorMagnitude(parseFloat(e.target.value))}
                        >
                    </div>

                    <div class="slider-control">
                        <div class="value-input-group">
                            <label>Angle</label>
                            <input type="number" class="input" step="1" min="0" max="360"
                                value={Math.round(getAngle(selectedVector))}
                                oninput={(e) => updateVectorAngle(parseFloat(e.target.value)||0)}
                            >
                            <span>°</span>
                        </div>
                        <input type="range" min="0" max="360" step="1"
                            value={getAngle(selectedVector)}
                            oninput={(e) => updateVectorAngle(parseFloat(e.target.value))}
                        >
                    </div>
                    
                    <div class="btn-group">
                        <button class="btn flex-1" onclick={duplicateVector}>Duplicate</button>
                        <button class="btn flex-1" onclick={deleteSelected}>Delete</button>
                    </div>
                 </div>
             </div>
        {/if}

        <SimilarityMatrix {vectors} />
    </div>

    <div class="controls-card card">
        <div class="control-section">
            <h3>Add Vector</h3>
            <p class="text-sm mb-4">Enter coordinates</p>
            <label class="label">Coordinates</label>
            <input type="text" class="input mb-4" placeholder="[x, y] or [[x], [y]]"
                bind:value={vectorInput}
                onkeypress={(e) => e.key === 'Enter' && addFromInput()}
            >
            <button class="btn btn-primary btn-full" onclick={addFromInput}>+ Add Vector</button>
        </div>

        {#if selectedVector}
             <div class="control-section desktop-vector-controls">
                <h3>Selected Vector</h3>
                <p class="text-sm mb-4">{selectedVector.name} = [{selectedVector.components.map(c => c.toFixed(1)).join(', ')}]</p>

                <div class="component-inputs">
                    <div class="component-input">
                        <label>X</label>
                        <input type="number" class="input" step="0.1" 
                            value={selectedVector.components[0]}
                            oninput={(e) => {
                                selectedVector.components[0] = parseFloat(e.target.value) || 0;
                                vectors = [...vectors];
                            }}
                        >
                    </div>
                    <div class="component-input">
                        <label>Y</label>
                        <input type="number" class="input" step="0.1" 
                            value={selectedVector.components[1]}
                            oninput={(e) => {
                                selectedVector.components[1] = parseFloat(e.target.value) || 0;
                                vectors = [...vectors];
                            }}
                        >
                    </div>
                </div>

                <div class="slider-control">
                    <div class="value-input-group">
                        <label>Magnitude</label>
                        <input type="number" class="input" step="0.1" min="0" 
                            value={getMagnitude(selectedVector).toFixed(1)}
                            oninput={(e) => updateVectorMagnitude(parseFloat(e.target.value)||0)}
                        >
                    </div>
                    <input type="range" min="0" max="20" step="0.1" 
                        value={Math.min(getMagnitude(selectedVector), 20)}
                        oninput={(e) => updateVectorMagnitude(parseFloat(e.target.value))}
                    >
                </div>

                <div class="slider-control">
                    <div class="value-input-group">
                        <label>Angle</label>
                        <input type="number" class="input" step="1" min="0" max="360"
                            value={Math.round(getAngle(selectedVector))}
                            oninput={(e) => updateVectorAngle(parseFloat(e.target.value)||0)}
                        >
                        <span>°</span>
                    </div>
                    <input type="range" min="0" max="360" step="1"
                        value={getAngle(selectedVector)}
                        oninput={(e) => updateVectorAngle(parseFloat(e.target.value))}
                    >
                </div>

                <div class="btn-group">
                    <button class="btn btn-sm flex-1" onclick={duplicateVector}>Duplicate</button>
                    <button class="btn btn-sm flex-1" onclick={deleteSelected}>Delete</button>
                </div>
            </div>
        {/if}

        <div class="control-section">
            <h3>All Vectors</h3>
            <div class="vector-list mt-4">
                {#each vectors as v (v.id)}
                    <div class="vector-item {selectedVector === v ? 'selected' : ''}" 
                         onclick={() => selectVector(v)} 
                         role="button" 
                         tabindex="0"
                         onkeydown={(e) => e.key === 'Enter' && selectVector(v)}
                    >
                        <div class="vector-color" style="background: {v.color}"></div>
                        <div class="vector-info">
                            <div class="vector-name">{v.name}</div>
                            <div class="vector-coords">[{v.components.map(c => c.toFixed(1)).join(', ')}]</div>
                        </div>
                        <button class="btn btn-sm" onclick={(e) => { e.stopPropagation(); deleteVector(v.id); }}>×</button>
                    </div>
                {/each}
            </div>
            <button class="btn btn-sm btn-full mt-4" onclick={clearAll}>Clear All</button>
        </div>
    </div>
</div>
