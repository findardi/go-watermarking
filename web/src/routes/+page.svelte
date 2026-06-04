<script lang="ts">
	import {
		requestWatermark,
		WatermarkError,
		zipResults,
		type Angle,
		type WatermarkConfig
	} from '$lib/api';

	let imageFiles = $state<File[]>([]);
	let watermarkFile = $state<File | null>(null);

	let markType = $state<'text' | 'image'>('text');
	let text = $state('CONFIDENTAL');
	let color = $state('#ffffff');
	let scale = $state(0.07);

	let placementMode = $state<'center' | 'pattern'>('center');
	let angle = $state<Angle>(0);
	let opacity = $state(0.5);

	let loading = $state(false);
	let zipping = $state(false);
	let errMsg = $state<string | null>(null);
	let results = $state<{ url: string; blob: Blob; format: string }[]>([]);
	let selectedIndex = $state(0);

	let canSubmit = $derived(
		imageFiles.length > 0 && (markType === 'text' || watermarkFile !== null)
	);

	function revokeResults() {
		for (const r of results) URL.revokeObjectURL(r.url);
		results = [];
	}

	function buildConfig(): WatermarkConfig {
		const mark =
			markType === 'text'
				? { type: 'text' as const, text, color, scale }
				: { type: 'image' as const, scale };

		return {
			mark,
			placement: { mode: placementMode, angle },
			opacity
		};
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		if (imageFiles.length === 0) return;

		loading = true;
		errMsg = null;
		revokeResults();

		try {
			const out = await requestWatermark(buildConfig(), {
				image: imageFiles,
				watermark: watermarkFile ?? undefined
			});

			results = out.map((r) => ({
				url: URL.createObjectURL(r.blob),
				blob: r.blob,
				format: r.format
			}));
			selectedIndex = 0;
		} catch (error) {
			errMsg = error instanceof WatermarkError ? error.message : 'failed to processing image';
		} finally {
			loading = false;
		}
	}

	async function downloadOne(i: number) {
		const r = results[i];
		const a = document.createElement('a');
		a.href = r.url;
		a.download = `watermark-${i + 1}.${r.format}`;
		a.click();
	}

	async function downloadZip() {
		if (results.length === 0) return;

		zipping = true;

		try {
			const blob = await zipResults(
				results.map((r) => ({
					blob: r.blob,
					format: r.format
				}))
			);

			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'watermark.zip';
			a.click();
			URL.revokeObjectURL(url);
		} finally {
			zipping = false;
		}
	}
</script>

<main class="min-h-screen bg-base-200 p-4 md:p-8">
	<div class="mx-auto max-w-5xl">
		<h1 class="mb-6 text-3xl font-bold">Watermarking</h1>

		<div class="grid gap-6 md:grid-cols-2">
			<!-- Configuration -->
			<div class="card bg-base-100 shadow">
				<div class="card-body">
					<h2 class="card-title">Configuration</h2>

					<form class="grid gap-4" onsubmit={handleSubmit}>
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Base Files</legend>
							<input
								type="file"
								accept="image/png,image/jpeg"
								multiple
								class="file-input w-full"
								onchange={(e) => (imageFiles = Array.from(e.currentTarget.files ?? []))}
							/>
						</fieldset>

						<fieldset class="fieldset">
							<legend class="fieldset-legend">Mark Type</legend>
							<select class="select w-full" bind:value={markType}>
								<option value="text">Text</option>
								<option value="image">Logo</option>
							</select>
						</fieldset>

						{#if markType === 'text'}
							<label class="floating-label">
								<span>Text</span>
								<input type="text" class="input w-full" bind:value={text} />
							</label>
							<label class="flex items-center gap-3">
								<span class="grow">Warna</span>
								<input type="color" class="h-10 w-14 rounded" bind:value={color} />
							</label>
						{:else}
							<fieldset class="fieldset">
								<legend class="fieldset-legend">Logo File</legend>
								<input
									type="file"
									accept="image/png,image/jpeg"
									class="file-input w-full"
									onchange={(e) => (watermarkFile = e.currentTarget.files?.[0] ?? null)}
								/>
							</fieldset>
						{/if}

						<div>
							<div class="mb-1 flex justify-between text-sm">
								<span>Size (Width Fraction)</span><span class="font-mono">{scale}</span>
							</div>
							<input
								type="range"
								min="0.02"
								max="0.5"
								step="0.01"
								class="range w-full"
								bind:value={scale}
							/>
						</div>

						<fieldset class="fieldset">
							<legend class="fieldset-legend">Placement</legend>
							<div class="flex gap-4">
								<label class="flex items-center gap-2">
									<input type="radio" class="radio" value="center" bind:group={placementMode} /> Center
								</label>
								<label class="flex items-center gap-2">
									<input type="radio" class="radio" value="pattern" bind:group={placementMode} /> Pattern
								</label>
							</div>
						</fieldset>

						{#if placementMode === 'pattern'}
							<fieldset class="fieldset">
								<legend class="fieldset-legend">Corner</legend>
								<select class="select w-full" bind:value={angle}>
									<option value={0}>0° (Straight)</option>
									<option value={45}>45° (Diagonal)</option>
								</select>
							</fieldset>
						{/if}

						<!-- Opacity -->
						<div>
							<div class="mb-1 flex justify-between text-sm">
								<span>Opacity</span><span class="font-mono">{opacity}</span>
							</div>
							<input
								type="range"
								min="0"
								max="1"
								step="0.05"
								class="range w-full"
								bind:value={opacity}
							/>
						</div>

						<button type="submit" class="btn btn-outline" disabled={!canSubmit || loading}>
							{#if loading}<span class="loading loading-spinner loading-sm"></span>{/if}
							{loading ? 'Processing..' : 'Create watermark'}
						</button>
					</form>
				</div>
			</div>

			<!-- Preview -->
			<div class="card bg-base-100 shadow">
				<div class="card-body">
					<h2 class="card-title">Preview</h2>

					{#if errMsg}
						<div role="alert" class="alert alert-error">
							<span>{errMsg}</span>
						</div>
					{/if}

					{#if loading}
						<div class="grid h-64 place-items-center">
							<span class="loading loading-spinner loading-lg"></span>
						</div>
					{:else if results.length > 0}
						<img src={results[selectedIndex].url} alt="Hasil watermark" class="w-full rounded" />

						{#if results.length > 1}
							<div class="mt-2 flex gap-2 overflow-x-auto pb-2">
								{#each results as r, i (r.url)}
									<button
										type="button"
										class="shrink-0 overflow-hidden rounded border {i === selectedIndex
											? 'ring-2 ring-neutral'
											: 'border-base-300'}"
										onclick={() => (selectedIndex = i)}
									>
										<img src={r.url} alt={`Hasil ${i + 1}`} class="h-16 w-16 object-cover" />
									</button>
								{/each}
							</div>
						{/if}

						<div class="mt-2 flex gap-2">
							<button
								class="btn btn-outline"
								onclick={() => {
									downloadOne(selectedIndex);
								}}
							>
								Download
							</button>
							<button
								type="button"
								class="btn btn-neutral"
								onclick={downloadZip}
								disabled={zipping}
							>
								{zipping ? 'Zipping..' : 'Download All (ZIP)'}
							</button>
						</div>
					{:else}
						<div
							class="grid h-64 place-items-center rounded border-2 border-dashed border-base-300 text-base-content/50"
						>
							Result Watermark
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</main>
