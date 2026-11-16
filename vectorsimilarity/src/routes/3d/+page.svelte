<script>
    import { VectorMath } from '$lib/vector-math.js';
    import Canvas3D from '$lib/components/Canvas3D.svelte';
    import SimilarityMatrix from '$lib/components/SimilarityMatrix.svelte';

    let vectors = $state([]);
    let selectedVector = $state(null);
    let vectorCounter = 0;
    const colors = ['#ef4444', '#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#06b6d4', '#84cc16'];
    
    let vectorInput = $state('');

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
        if (!parsed || parsed.length < 3 || parsed.some(isNaN)) {
            alert('Invalid format. Use [[x], [y], [z]]');
            return;
        }
        addVector([parsed[0], parsed[1], parsed[2]]);
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
    
    function clearAll() {
        vectors = [];
        selectedVector = null;
    }

</script>

<div class="main-layout">
    <div>
        <div class="card mb-4">
            <div class="canvas-wrapper">
                <div class="canvas-container w-full h-full">
                    <Canvas3D 
                        {vectors} 
                        {selectedVector} 
                        onselect={selectVector} 
                    />
                </div>
            </div>
        </div>

        <SimilarityMatrix {vectors} />
    </div>

    <div class="controls-card">
        <div class="card mb-4">
            <div class="card-header">
                <h2>Add Vector</h2>
            </div>
            <div class="card-body">
                <label class="label">Enter coordinates</label>
                <input type="text" class="input mb-4" placeholder="[[x], [y], [z]]"
                    bind:value={vectorInput}
                    onkeypress={(e) => e.key === 'Enter' && addFromInput()}
                >
                <button class="btn btn-primary" style="width: 100%" onclick={addFromInput}>Add Vector</button>
                <p class="text-sm text-center mt-4">
                    Format: [[x], [y], [z]]
                </p>
            </div>
        </div>

        <div class="card mb-4">
            <div class="card-header">
                <h2>Vectors</h2>
            </div>
            <div class="card-body">
                <div class="vector-list">
                    {#each vectors as v}
                        <div class="vector-item">
                            <div class="vector-color" style="background: {v.color}"></div>
                            <div class="vector-info">
                                <div class="vector-name">{v.name}</div>
                                <div class="vector-coords">[{v.components.map(c => c.toFixed(1)).join(', ')}]</div>
                            </div>
                            <button class="btn btn-sm" onclick={() => deleteVector(v.id)}>×</button>
                        </div>
                    {/each}
                </div>
                <button class="btn btn-sm mt-4" ID="clearAllBtn" onclick={clearAll}>Clear All</button>
            </div>
        </div>

        {#if selectedVector}
            <div class="card">
                <div class="card-header">
                    <h2>Selected Vector</h2>
                </div>
                <div class="card-body">
                    <strong>{selectedVector.name}</strong> = [{selectedVector.components.map(c => c.toFixed(1)).join(', ')}]<br><br>
                    Magnitude: {VectorMath.magnitude(selectedVector.components).toFixed(3)}<br><br>
                    {#if vectors.length > 1}
                        <strong>Similarities:</strong><br>
                        {#each vectors as v}
                            {#if v.id !== selectedVector.id}
                                {@const sim = VectorMath.cosineSimilarity(selectedVector.components, v.components)}
                                {v.name}: {sim.toFixed(3)} ({(sim * 100).toFixed(1)}%)<br>
                            {/if}
                        {/each}
                    {/if}
                </div>
            </div>
        {/if}
    </div>
</div>
