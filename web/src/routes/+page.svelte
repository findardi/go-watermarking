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
			placement: {
				mode: placementMode,
				angle: angle
			},
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

<main>
	<h1>Watermarking</h1>

	<form onsubmit={handleSubmit}>
		<!-- Base image (wajib) -->
		<label>
			Gambar dasar
			<input
				type="file"
				accept="image/png,image/jpeg"
				onchange={(e) => (imageFile = e.currentTarget.files?.[0] ?? null)}
			/>
		</label>

		<!-- Jenis mark -->
		<fieldset>
			<legend>Jenis watermark</legend>
			<label><input type="radio" value="text" bind:group={markType} /> Teks</label>
			<label><input type="radio" value="image" bind:group={markType} /> Logo</label>
		</fieldset>

		<!-- Field khusus mark teks -->
		{#if markType === 'text'}
			<label>
				Teks
				<input type="text" bind:value={text} />
			</label>
			<label>
				Warna
				<input type="color" bind:value={color} />
			</label>
		{:else}
			<label>
				File Logo
				<input
					type="file"
					accept="image/png,image/jpeg"
					onchange={(e) => (watermarkFile = e.currentTarget.files?.[0] ?? null)}
				/>
			</label>
		{/if}

		<!-- Scale: dipakai kedua jenis -->
		<label>
			Ukuran (fraksi lebar gambar): {scale}
			<input type="range" min="0.02" max="0.5" step="0.01" bind:value={scale} />
		</label>

		<!-- Penempatan -->
		<fieldset>
			<legend>Penempatan</legend>
			<label><input type="radio" value="center" bind:group={placementMode} /> Tengah</label>
			<label><input type="radio" value="pattern" bind:group={placementMode} /> Pola b</label>
		</fieldset>

		{#if placementMode === 'pattern'}
			<label>
				Sudut
				<select bind:value={angle}>
					<option value={0}>0° (lurus)</option>
					<option value={45}>45° (diagonal)</option>
				</select>
			</label>
		{/if}

		<!-- Opacity -->
		<label>
			Opacity: {opacity}
			<input type="range" min="0" max="1" step="0.05" bind:value={opacity} />
		</label>

		<button type="submit" disabled={!canSubmit || loading}>
			{loading ? 'Memproses…' : 'Buat watermark'}
		</button>
	</form>

	{#if errMsg}
		<p class="error" role="alert">{errMsg}</p>
	{/if}

	{#if resUrl}
		<section>
			<h2>Hasil</h2>
			<img src={resUrl} alt="Hasil watermark" />
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- blob object URL rute -->
			<a href={resUrl} download="watermarked">Download</a>
		</section>
	{/if}
</main>

<style>
	main {
		max-width: 32rem;
		margin: 2rem auto;
		display: grid;
		gap: 1rem;
		font-family: system-ui, sans-serif;
	}
	form {
		display: grid;
		gap: 0.75rem;
	}
	label {
		display: grid;
		gap: 0.25rem;
	}
	fieldset {
		display: grid;
		gap: 0.25rem;
	}
	.error {
		color: #b00020;
	}
	img {
		max-width: 100%;
		height: auto;
	}
</style>
