<script>
    import { VectorMath } from '$lib/vector-math.js';

    let { vectors = [] } = $props();

    function getSimilarityClass(sim) {
        const absSim = Math.abs(sim);
        if (absSim >= 0.8) return 'similarity-high';
        if (absSim >= 0.5) return 'similarity-medium';
        return 'similarity-low';
    }
</script>

<div class="card">
    <div class="card-header">
        <h2>Similarity Matrix</h2>
    </div>
    <div class="card-body">
        {#if vectors.length < 2}
            <p class="text-sm text-center">Add at least 2 vectors</p>
        {:else}
            <div style="overflow-x: auto;">
                <table class="table">
                    <thead>
                        <tr>
                            <th></th>
                            {#each vectors as v}
                                <th>{v.name}</th>
                            {/each}
                        </tr>
                    </thead>
                    <tbody>
                        {#each vectors as v1}
                            <tr>
                                <th>{v1.name}</th>
                                {#each vectors as v2}
                                    {@const sim = VectorMath.cosineSimilarity(v1.components, v2.components)}
                                    <td>
                                        <span class={getSimilarityClass(sim)}>
                                            {sim.toFixed(3)}
                                        </span>
                                    </td>
                                {/each}
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>
        {/if}
    </div>
</div>
