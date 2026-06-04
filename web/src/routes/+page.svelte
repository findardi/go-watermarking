<script lang="ts">
	import { requestWatermark, WatermarkError, type Angle, type WatermarkConfig } from '$lib/api';

	let imageFile = $state<File | null>(null);
	let watermarkFile = $state<File | null>(null);

	let markType = $state<'text' | 'image'>('text');
	let text = $state('watermark');
	let color = $state('#ffffff');
	let scale = $state(0.07);

	let placementMode = $state<'center' | 'pattern'>('center');
	let angle = $state<Angle>(0);
	let opacity = $state(0.5);

	let loading = $state(false);
	let errMsg = $state<string | null>(null);
	let resUrl = $state<string | null>(null);

	let canSubmit = $derived(imageFile !== null && (markType === 'text' || watermarkFile !== null));

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
		if (!imageFile) return;

		loading = true;
		errMsg = null;

		if (resUrl) {
			URL.revokeObjectURL(resUrl);
			resUrl = null;
		}

		try {
			const blob = await requestWatermark(buildConfig(), {
				image: imageFile,
				watermark: watermarkFile ?? undefined
			});
			resUrl = URL.createObjectURL(blob);
		} catch (error) {
			errMsg = error instanceof WatermarkError ? error.message : 'failed to processing image';
		} finally {
			loading = false;
		}
	}
</script>

<main class="min-h-screen bg-base-200 p-4 md:p-8">
	<div class="mx-auto max-w-5xl">
		<h1 class="mb-6 text-3xl font-bold">Watermarking</h1>

		<div class="grid gap-6 md:grid-cols-2">
			<!-- KIRI: Konfigurasi -->
			<div class="card bg-base-100 shadow">
				<div class="card-body">
					<h2 class="card-title">Konfigurasi</h2>

					<form class="grid gap-4" onsubmit={handleSubmit}>
						<!-- Gambar dasar -->
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Gambar dasar</legend>
							<input
								type="file"
								accept="image/png,image/jpeg"
								class="file-input w-full"
								onchange={(e) => (imageFile = e.currentTarget.files?.[0] ?? null)}
							/>
						</fieldset>

						<!-- Jenis mark -->
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Jenis watermark</legend>
							<select class="select w-full" bind:value={markType}>
								<option value="text">Teks</option>
								<option value="image">Logo</option>
							</select>
						</fieldset>

						<!-- Field khusus -->
						{#if markType === 'text'}
							<label class="floating-label">
								<span>Teks</span>
								<input type="text" class="input w-full" bind:value={text} />
							</label>
							<label class="flex items-center gap-3">
								<span class="grow">Warna</span>
								<input type="color" class="h-10 w-14 rounded" bind:value={color} />
							</label>
						{:else}
							<fieldset class="fieldset">
								<legend class="fieldset-legend">File Logo</legend>
								<input
									type="file"
									accept="image/png,image/jpeg"
									class="file-input w-full"
									onchange={(e) => (watermarkFile = e.currentTarget.files?.[0] ?? null)}
								/>
							</fieldset>
						{/if}

						<!-- Scale -->
						<div>
							<div class="mb-1 flex justify-between text-sm">
								<span>Ukuran (fraksi lebar)</span><span class="font-mono">{scale}</span>
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

						<!-- Penempatan -->
						<fieldset class="fieldset">
							<legend class="fieldset-legend">Penempatan</legend>
							<div class="flex gap-4">
								<label class="flex items-center gap-2">
									<input type="radio" class="radio" value="center" bind:group={placementMode} /> Tengah
								</label>
								<label class="flex items-center gap-2">
									<input type="radio" class="radio" value="pattern" bind:group={placementMode} /> Pola
								</label>
							</div>
						</fieldset>

						{#if placementMode === 'pattern'}
							<fieldset class="fieldset">
								<legend class="fieldset-legend">Sudut</legend>
								<select class="select w-full" bind:value={angle}>
									<option value={0}>0° (lurus)</option>
									<option value={45}>45° (diagonal)</option>
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

						<button type="submit" class="btn btn-primary" disabled={!canSubmit || loading}>
							{#if loading}<span class="loading loading-spinner loading-sm"></span>{/if}
							{loading ? 'Memproses…' : 'Buat watermark'}
						</button>
					</form>
				</div>
			</div>

			<!-- KANAN: Preview -->
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
					{:else if resUrl}
						<img src={resUrl} alt="Hasil watermark" class="w-full rounded" />
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a href={resUrl} download="watermarked" class="btn btn-outline mt-2">Download</a>
					{:else}
						<div
							class="grid h-64 place-items-center rounded border-2 border-dashed border-base-300 text-base-content/50"
						>
							Hasil watermark akan muncul di sini
						</div>
					{/if}
				</div>
			</div>
		</div>
	</div>
</main>
