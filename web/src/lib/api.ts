import JSZip from 'jszip';

const API_BASE = 'http://localhost:8080';
const ENDPOINT = `${API_BASE}/api/v1/watermark`;

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

interface ResultBody {
	data: string;
	format: string;
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

export async function parseWatermarkResponse(resp: Response): Promise<WatermarkResult[]> {
	if (resp.ok) {
		const body = (await resp.json()) as ResultBody[];
		return body.map((r) => ({
			blob: base64ToBlob(r.data, formatToMime(r.format)),
			format: r.format
		}));
	}

	let code = resp.status;
	let message = `${resp.status} ${resp.statusText}`;

	try {
		const body = (await resp.json()) as ErrorBody;
		if (body?.error?.message) {
			code = body.error.code;
			message = body.error.message;
		}
	} catch {
		// using fallback status in above
	}

	throw new WatermarkError(code, message);
}

export async function zipResults(result: WatermarkResult[]): Promise<Blob> {
	const zip = new JSZip();
	result.forEach((r, i) => {
		zip.file(`watermark-${i + 1}.${r.format}`, r.blob);
	});

	return zip.generateAsync({ type: 'blob' });
}

export async function requestWatermark(
	config: WatermarkConfig,
	files: WatermarkFiles
): Promise<WatermarkResult[]> {
	const form = buildWatermarkform(config, files);
	const resp = await fetch(ENDPOINT, {
		method: 'POST',
		body: form
	});

	return parseWatermarkResponse(resp);
}
