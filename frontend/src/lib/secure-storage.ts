const DATABASE_NAME = "anochat-secure-storage";
const DATABASE_VERSION = 1;
const KEY_STORE_NAME = "keys";
const DEVICE_KEY_ID = "device-key-v1";
const STORAGE_PREFIX = "anochat:secure:";

interface EncryptedEnvelope {
	v: 1;
	alg: "AES-GCM";
	iv: string;
	ciphertext: string;
}

let databasePromise: Promise<IDBDatabase> | null = null;
let deviceKeyPromise: Promise<CryptoKey> | null = null;

function openDatabase(): Promise<IDBDatabase> {
	if (databasePromise) return databasePromise;

	databasePromise = new Promise((resolve, reject) => {
		const request = indexedDB.open(DATABASE_NAME, DATABASE_VERSION);
		request.onupgradeneeded = () => {
			if (!request.result.objectStoreNames.contains(KEY_STORE_NAME)) {
				request.result.createObjectStore(KEY_STORE_NAME);
			}
		};
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
		request.onblocked = () =>
			reject(new Error("Secure storage database is blocked"));
	});

	return databasePromise;
}

async function readDeviceKey(): Promise<CryptoKey | undefined> {
	const database = await openDatabase();
	return new Promise((resolve, reject) => {
		const request = database
			.transaction(KEY_STORE_NAME, "readonly")
			.objectStore(KEY_STORE_NAME)
			.get(DEVICE_KEY_ID);
		request.onsuccess = () => resolve(request.result as CryptoKey | undefined);
		request.onerror = () => reject(request.error);
	});
}

async function addDeviceKey(key: CryptoKey): Promise<void> {
	const database = await openDatabase();
	return new Promise((resolve, reject) => {
		const request = database
			.transaction(KEY_STORE_NAME, "readwrite")
			.objectStore(KEY_STORE_NAME)
			.add(key, DEVICE_KEY_ID);
		request.onsuccess = () => resolve();
		request.onerror = () => reject(request.error);
	});
}

async function getDeviceKey(): Promise<CryptoKey> {
	if (deviceKeyPromise) return deviceKeyPromise;

	deviceKeyPromise = (async () => {
		const existingKey = await readDeviceKey();
		if (existingKey) return existingKey;

		const generatedKey = await crypto.subtle.generateKey(
			{ name: "AES-GCM", length: 256 },
			false,
			["encrypt", "decrypt"],
		);

		try {
			await addDeviceKey(generatedKey);
			return generatedKey;
		} catch (error) {
			// Another tab may have created the key after our initial read.
			if (error instanceof DOMException && error.name === "ConstraintError") {
				const concurrentKey = await readDeviceKey();
				if (concurrentKey) return concurrentKey;
			}
			throw error;
		}
	})();

	try {
		return await deviceKeyPromise;
	} catch (error) {
		deviceKeyPromise = null;
		throw error;
	}
}

function bytesToBase64(bytes: Uint8Array): string {
	let binary = "";
	for (const byte of bytes) binary += String.fromCharCode(byte);
	return btoa(binary);
}

function base64ToBytes(value: string): Uint8Array<ArrayBuffer> {
	const binary = atob(value);
	const bytes = new Uint8Array(binary.length);
	for (let index = 0; index < binary.length; index += 1) {
		bytes[index] = binary.charCodeAt(index);
	}
	return bytes;
}

function isEncryptedEnvelope(value: unknown): value is EncryptedEnvelope {
	if (!value || typeof value !== "object") return false;
	const envelope = value as Partial<EncryptedEnvelope>;
	return (
		envelope.v === 1 &&
		envelope.alg === "AES-GCM" &&
		typeof envelope.iv === "string" &&
		typeof envelope.ciphertext === "string"
	);
}

function storageKey(key: string): string {
	return `${STORAGE_PREFIX}${key}`;
}

export const secureStorage = {
	async get<T>(key: string): Promise<T | null> {
		if (
			typeof window === "undefined" ||
			!window.indexedDB ||
			!window.crypto?.subtle
		) {
			return null;
		}

		try {
			const stored = localStorage.getItem(storageKey(key));
			if (!stored) return null;
			const envelope: unknown = JSON.parse(stored);
			if (!isEncryptedEnvelope(envelope)) return null;

			const plaintext = await crypto.subtle.decrypt(
				{
					name: "AES-GCM",
					iv: base64ToBytes(envelope.iv),
					additionalData: new TextEncoder().encode(storageKey(key)),
				},
				await getDeviceKey(),
				base64ToBytes(envelope.ciphertext),
			);
			return JSON.parse(new TextDecoder().decode(plaintext)) as T;
		} catch {
			return null;
		}
	},

	async set<T>(key: string, value: T): Promise<boolean> {
		if (
			typeof window === "undefined" ||
			!window.indexedDB ||
			!window.crypto?.subtle
		) {
			return false;
		}

		try {
			const iv = crypto.getRandomValues(new Uint8Array(12));
			const ciphertext = await crypto.subtle.encrypt(
				{
					name: "AES-GCM",
					iv,
					additionalData: new TextEncoder().encode(storageKey(key)),
				},
				await getDeviceKey(),
				new TextEncoder().encode(JSON.stringify(value)),
			);
			const envelope: EncryptedEnvelope = {
				v: 1,
				alg: "AES-GCM",
				iv: bytesToBase64(iv),
				ciphertext: bytesToBase64(new Uint8Array(ciphertext)),
			};
			localStorage.setItem(storageKey(key), JSON.stringify(envelope));
			return true;
		} catch {
			return false;
		}
	},

	remove(key: string): void {
		if (typeof window === "undefined") return;
		try {
			localStorage.removeItem(storageKey(key));
		} catch {}
	},
};
