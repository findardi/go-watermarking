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
	image: File;
	watermark?: File;
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

export function buildWatermarkform(config: WatermarkConfig, files: WatermarkFiles): FormData {
	const form = new FormData();
	form.append('config', JSON.stringify(config));
	form.append('image', files.image);

	if (config.mark.type === 'image') {
		if (!files.watermark) {
			throw new WatermarkError(0, 'watermark file is required');
		}
		form.append('watermark', files.watermark);
	}

	return form;
}

export async function parseWatermarkResponse(resp: Response): Promise<Blob> {
	if (resp.ok) {
		return resp.blob();
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

export async function requestWatermark(
	config: WatermarkConfig,
	files: WatermarkFiles
): Promise<Blob> {
	const form = buildWatermarkform(config, files);
	const resp = await fetch(ENDPOINT, {
		method: 'POST',
		body: form
	});

	return parseWatermarkResponse(resp);
}
