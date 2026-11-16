<script>
    import '../app.css';
    import ThemeToggle from '$lib/components/ThemeToggle.svelte';
    import { page } from '$app/stores';
    import favicon from '$lib/assets/favicon.svg';

    let { children } = $props();
</script>

<svelte:head>
    <link rel="icon" href={favicon} />
</svelte:head>

<div class="container" style="flex: 1;">
    <header class="header">
        <h1>{$page.url.pathname === '/3d' ? '3D Cosine Similarity' : 'Cosine Similarity Calculator'}</h1>
        <nav class="nav">
            <ThemeToggle />
            {#if $page.url.pathname === '/3d'}
                <a href="/" class="btn">Home</a>
                <a href="/" class="btn">2D Mode</a>
            {:else}
                <a href="/3d" class="btn">3D Mode</a>
            {/if}
        </nav>
    </header>

    {@render children()}
</div>
