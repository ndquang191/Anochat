"use client";

import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { pushAPI } from "@/lib/api";
import { useLanguage } from "@/contexts/theme";
import { useInvalidateUserState } from "@/hooks/queries/use-user-state";
import { getCookie } from "@/lib/cookies";

const SUBSCRIPTION_ID_KEY = "anochat:push-subscription-id";

interface BeforeInstallPromptEvent extends Event {
	prompt(): Promise<void>;
	userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

interface PWAContextValue {
	isOnline: boolean;
	isStandalone: boolean;
	isIOS: boolean;
	canInstall: boolean;
	pushEnabled: boolean;
	isPushSubscribed: boolean;
	install(): Promise<boolean>;
	refreshPushConfig(): Promise<void>;
	ensurePushSubscription(): Promise<string | null>;
	unsubscribe(): Promise<void>;
}

const PWAContext = createContext<PWAContextValue | null>(null);

function decodeApplicationServerKey(value: string): Uint8Array<ArrayBuffer> {
	const padding = "=".repeat((4 - value.length % 4) % 4);
	const base64 = (value + padding).replace(/-/g, "+").replace(/_/g, "/");
	const raw = atob(base64);
	return Uint8Array.from(raw, (character) => character.charCodeAt(0));
}

export function PWAProvider({ children }: { children: ReactNode }) {
	const { language } = useLanguage();
	const invalidateUserState = useInvalidateUserState();
	const [isOnline, setIsOnline] = useState(true);
	const [installPrompt, setInstallPrompt] = useState<BeforeInstallPromptEvent | null>(null);
	const [isStandalone, setIsStandalone] = useState(false);
	const [isIOS, setIsIOS] = useState(false);
	const [pushConfig, setPushConfig] = useState<{ enabled: boolean; public_key: string } | null>(null);
	const [isPushSubscribed, setIsPushSubscribed] = useState(false);

	useEffect(() => {
		setIsOnline(navigator.onLine);
		setIsStandalone(window.matchMedia("(display-mode: standalone)").matches || (navigator as Navigator & { standalone?: boolean }).standalone === true);
		setIsIOS(/iphone|ipad|ipod/i.test(navigator.userAgent));
		const online = () => setIsOnline(true);
		const offline = () => setIsOnline(false);
		const beforeInstall = (event: Event) => {
			event.preventDefault();
			setInstallPrompt(event as BeforeInstallPromptEvent);
		};
		window.addEventListener("online", online);
		window.addEventListener("offline", offline);
		window.addEventListener("beforeinstallprompt", beforeInstall);
		if ("serviceWorker" in navigator) {
			navigator.serviceWorker.register("/sw.js", { scope: "/", updateViaCache: "none" })
				.then((registration) => registration.pushManager.getSubscription())
				.then((subscription) => setIsPushSubscribed(!!subscription))
				.catch(console.error);
			navigator.serviceWorker.addEventListener("message", (event) => {
				if (event.data?.type === "match_found") invalidateUserState();
			});
		}
		if (getCookie("has_session")) {
			pushAPI.config().then((response) => setPushConfig(response.data ?? null)).catch(() => setPushConfig(null));
		}
		return () => {
			window.removeEventListener("online", online);
			window.removeEventListener("offline", offline);
			window.removeEventListener("beforeinstallprompt", beforeInstall);
		};
	}, [invalidateUserState]);

	const install = useCallback(async () => {
		if (!installPrompt) return false;
		await installPrompt.prompt();
		const choice = await installPrompt.userChoice;
		setInstallPrompt(null);
		return choice.outcome === "accepted";
	}, [installPrompt]);

	const refreshPushConfig = useCallback(async () => {
		try {
			const response = await pushAPI.config();
			setPushConfig(response.data ?? null);
		} catch {
			setPushConfig(null);
		}
	}, []);

	const ensurePushSubscription = useCallback(async () => {
		if (!("serviceWorker" in navigator) || !("PushManager" in window) || !("Notification" in window)) return null;
		if (!pushConfig?.enabled || !pushConfig.public_key) return null;
		let permission = Notification.permission;
		if (permission === "default") permission = await Notification.requestPermission();
		if (permission !== "granted") return null;
		const registration = await navigator.serviceWorker.ready;
		let subscription = await registration.pushManager.getSubscription();
		if (!subscription) {
			subscription = await registration.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: decodeApplicationServerKey(pushConfig.public_key),
			});
		}
		const json = subscription.toJSON();
		if (!json.endpoint || !json.keys?.p256dh || !json.keys.auth) return null;
		const response = await pushAPI.subscribe({ endpoint: json.endpoint, keys: { p256dh: json.keys.p256dh, auth: json.keys.auth }, locale: language });
		const id = response.data?.id ?? null;
		if (id) localStorage.setItem(SUBSCRIPTION_ID_KEY, id);
		setIsPushSubscribed(!!id);
		return id;
	}, [language, pushConfig]);

	useEffect(() => {
		if (!pushConfig?.enabled || !isPushSubscribed || Notification.permission !== "granted") return;
		ensurePushSubscription().catch(() => {});
	}, [language, pushConfig?.enabled, isPushSubscribed, ensurePushSubscription]);

	const unsubscribe = useCallback(async () => {
		const id = localStorage.getItem(SUBSCRIPTION_ID_KEY);
		if (id) { try { await pushAPI.unsubscribe(id); } catch {} }
		if ("serviceWorker" in navigator) {
			try { const registration = await navigator.serviceWorker.ready; await (await registration.pushManager.getSubscription())?.unsubscribe(); } catch {}
		}
		localStorage.removeItem(SUBSCRIPTION_ID_KEY);
		setIsPushSubscribed(false);
		invalidateUserState();
	}, [invalidateUserState]);

	const value = useMemo(() => ({ isOnline, isStandalone, isIOS, canInstall: !!installPrompt || (isIOS && !isStandalone), pushEnabled: pushConfig?.enabled === true, isPushSubscribed, install, refreshPushConfig, ensurePushSubscription, unsubscribe }), [isOnline, isStandalone, isIOS, installPrompt, pushConfig?.enabled, isPushSubscribed, install, refreshPushConfig, ensurePushSubscription, unsubscribe]);
	return <PWAContext.Provider value={value}>{!isOnline && <div className="fixed inset-x-0 top-0 z-[10000] bg-amber-500 px-3 py-1 text-center text-xs font-medium text-black">{language === "vi" ? "Mất kết nối mạng — AnoChat sẽ tự kết nối lại." : "You’re offline — AnoChat will reconnect automatically."}</div>}{children}</PWAContext.Provider>;
}

export function usePWA() {
	const value = useContext(PWAContext);
	if (!value) throw new Error("usePWA must be used inside PWAProvider");
	return value;
}

export function getStoredPushSubscriptionID(): string | null {
	if (typeof window === "undefined") return null;
	return localStorage.getItem(SUBSCRIPTION_ID_KEY);
}
