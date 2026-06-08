import JSZip from 'jszip';

const API_BASE = 'http://localhost:8080';
const ENDPOINT = `${API_BASE}/api/v1/watermark`;
const SAFE_BATCH_BYTES = 180 * 1024 * 1024;
const MAX_FILES_BYTES = 20 * 1024 * 1024;

export type Angle = 0 | 45;

export type Mark =
	| { type: 'text'; text: string; color: string; scale: number }
	| { type: 'image'; scale: number };

export interface WatermarkConfig {
	mark: Mark;
	placement: { mode: 'center' | 'pattern'; angle: Angle };
	opacity: number;
}

export interface WatermarkFiles {
	image: File[];
	watermark?: File;
}

export interface WatermarkResult {
	blob: Blob;
	format: string;
}

export class WatermarkError extends Error {
	constructor(
		public readonly code: number,
		message: string
	) {
		super(message);
		this.name = 'watermarkError';
	}
}

interface ErrorBody {
	error: {
		code: number;
		message: string;
	};
}

function formatToMime(format: string): string {
	switch (format) {
		case 'png':
			return 'image/png';
		case 'jpeg':
			return 'image/jpeg';
		default:
			return 'application/octet-stream';
	}
}

function base64ToBlob(base64: string, mime: string): Blob {
	const binary = atob(base64);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) {
		bytes[i] = binary.charCodeAt(i);
	}
	return new Blob([bytes], { type: mime });
}

export function buildWatermarkform(config: WatermarkConfig, files: WatermarkFiles): FormData {
	const form = new FormData();
	form.append('config', JSON.stringify(config));

	for (const image of files.image) {
		form.append('image', image);
	}

	if (config.mark.type === 'image') {
		if (!files.watermark) {
			throw new WatermarkError(0, 'watermark file is required');
		}
		form.append('watermark', files.watermark);
	}

	return form;
}

export async function zipResults(result: WatermarkResult[]): Promise<Blob> {
	const zip = new JSZip();
	result.forEach((r, i) => {
		zip.file(`watermark-${i + 1}.${r.format}`, r.blob);
	});

	return zip.generateAsync({ type: 'blob' });
}

export interface BatchPlanOptions {
	maxBatchBytes: number;
	maxFileBytes: number;
	reservedBytes: number;
}

export interface RejectedFile {
	file: File;
	reason: 'too-large';
}

export interface BatchPlan {
	batches: File[][];
	rejected: RejectedFile[];
}

export function planBatches(files: File[], opts: BatchPlanOptions): BatchPlan {
	const batches: File[][] = [];
	const rejected: RejectedFile[] = [];

	const budget = opts.maxBatchBytes - opts.reservedBytes;

	let current: File[] = [];
	let currentBytes = 0;

	for (const file of files) {
		if (file.size > opts.maxFileBytes) {
			rejected.push({
				file,
				reason: 'too-large'
			});
			continue;
		}

		const exceedsBytes = currentBytes + file.size > budget;

		if (current.length > 0 && exceedsBytes) {
			batches.push(current);
			current = [];
			currentBytes = 0;
		}

		current.push(file);
		currentBytes += file.size;
	}

	if (current.length > 0) {
		batches.push(current);
	}

	return { batches, rejected };
}

export function batchOptions(config: WatermarkConfig, files: WatermarkFiles): BatchPlanOptions {
	const watermarkBytes = config.mark.type === 'image' && files.watermark ? files.watermark.size : 0;
	const configBytes = JSON.stringify(config).length;
	const overhead = 1024;
	return {
		maxBatchBytes: SAFE_BATCH_BYTES,
		maxFileBytes: MAX_FILES_BYTES,
		reservedBytes: watermarkBytes + configBytes + overhead
	};
}

interface ServerImageResult {
	index: number;
	filename?: string;
	status: 'ok' | 'error';
	format?: string;
	data?: string;
	error?: string;
}

interface ServerResponse {
	total: number;
	success: number;
	failed: number;
	results: ServerImageResult[];
}

export interface BatchProgress {
	processed: number;
	total: number;
	batchIndex: number;
	batchCount: number;
}

export interface ImageOutcome {
	index: number;
	filename?: string;
	status: 'ok' | 'error';
	blob?: Blob;
	format?: string;
	error?: string;
}

export interface BatchResult {
	total: number;
	success: number;
	failed: number;
	outcomes: ImageOutcome[];
}

async function throwResponseError(resp: Response): Promise<never> {
	let code = resp.status;
	let message = `${resp.status} ${resp.statusText}`;

	try {
		const body = (await resp.json()) as ErrorBody;
		if (body?.error?.message) {
			code = body.error.code;
			message = body.error.message;
		}
	} catch {
		// using fallback
	}
	throw new WatermarkError(code, message);
}

async function sendBatch(
	config: WatermarkConfig,
	watermark: File | undefined,
	images: File[]
): Promise<ServerImageResult[]> {
	const form = buildWatermarkform(config, {
		image: images,
		watermark
	});

	const resp = await fetch(ENDPOINT, {
		method: 'POST',
		body: form
	});

	if (!resp.ok) {
		await throwResponseError(resp);
	}

	const body = (await resp.json()) as ServerResponse;
	return body.results;
}

function toOutcome(r: ServerImageResult, globalIndex: number): ImageOutcome {
	if (r.status === 'ok' && r.data && r.format) {
		return {
			index: globalIndex,
			filename: r.filename,
			status: 'ok',
			blob: base64ToBlob(r.data, formatToMime(r.format)),
			format: r.format
		};
	}

	return {
		index: globalIndex,
		filename: r.filename,
		status: 'error',
		error: r.error ?? 'failed to process'
	};
}

export async function processInBatches(
	config: WatermarkConfig,
	files: WatermarkFiles,
	opts: BatchPlanOptions,
	onProgress: (p: BatchProgress) => void
): Promise<BatchResult> {
	const { batches, rejected } = planBatches(files.image, opts);

	const indexOf = new Map<File, number>();
	files.image.forEach((file, i) => indexOf.set(file, i));

	const outComes: ImageOutcome[] = [];

	for (const { file } of rejected) {
		outComes.push({
			index: indexOf.get(file) ?? -1,
			filename: file.name,
			status: 'error',
			error: 'file exceeds more than 20MB'
		});
	}

	const total = files.image.length;
	const batchCount = batches.length;
	let processed = rejected.length;

	onProgress({ processed, total, batchIndex: 0, batchCount });

	for (let b = 0; b < batches.length; b++) {
		const batch = batches[b];

		try {
			const results = await sendBatch(config, files.watermark, batch);
			for (const r of results) {
				const file = batch[r.index];
				outComes.push(toOutcome(r, indexOf.get(file) ?? -1));
			}
		} catch (err) {
			const message = err instanceof Error ? err.message : 'batch gagal';
			for (const file of batch) {
				outComes.push({
					index: indexOf.get(file) ?? -1,
					filename: file.name,
					status: 'error',
					error: message
				});
			}
		}

		processed += batch.length;
		onProgress({ processed, total, batchIndex: b + 1, batchCount });
	}

	outComes.sort((a, b) => a.index - b.index);

	const success = outComes.filter((o) => o.status === 'ok').length;
	const failed = outComes.length - success;

	return {
		total,
		success,
		failed,
		outcomes: outComes
	};
}
