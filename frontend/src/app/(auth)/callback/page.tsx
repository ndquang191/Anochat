"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/contexts/auth";
import { useLanguage } from "@/contexts/theme";
import { translateStored } from "@/lib/i18n";
import { toast } from "sonner";

export default function CallbackPage() {
	const router = useRouter();
	const { login } = useAuth();
	const { t } = useLanguage();

	useEffect(() => {
		let active = true;
		void login()
			.then(() => {
				if (active) router.replace("/");
			})
			.catch((error) => {
				if (!active) return;
				toast.error(error instanceof Error ? error.message : translateStored("apiUnavailable"));
				router.replace("/login");
			});
		return () => {
			active = false;
		};
	}, [login, router]);

	return (
		<div className="flex min-h-svh w-full items-center justify-center p-6">
			<div className="text-center">
				<div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4" />
				<h1 className="text-2xl font-bold mb-2">{t("processing")}</h1>
				<p className="text-gray-600">{t("loading")}</p>
			</div>
		</div>
	);
}
