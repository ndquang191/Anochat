"use client";

import React, { useState } from "react";
import { Flag, X } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useAlertDialogContext } from "@/contexts/alert-dialog";
import { useLanguage } from "@/contexts/theme";
import { AppActionButton } from "@/components/header/app-action-button";
import { moderationAPI } from "@/lib/api";
import { toast } from "sonner";

interface HeaderProps {
	trigger: React.ReactNode;
}

export default function AppHeader({ trigger }: HeaderProps) {
	const { room } = useAuth();
	const { isAdminOpen, setIsAdminOpen } = useAdmin();
	const { t } = useLanguage();
	const partner = room?.partner;
	const [reported, setReported] = useState(false);
	const alertDialog = useAlertDialogContext();

	const handleReport = async () => {
		if (!partner || !room || reported) return;

		const confirmed = await alertDialog.open({
			title: t("reportConfirmTitle"),
			description: t("reportConfirmDescription"),
			confirmText: t("report"),
			cancelText: t("cancel"),
		});

		if (!confirmed) return;

		try {
			await moderationAPI.createReport(partner.id, room.id);
			setReported(true);
			toast.success(t("reportSubmitted"));
		} catch {
			// error toast handled by apiCall
		}
	};

	const isHidden = partner?.profile?.is_hidden;
	const partnerName = isHidden ? t("anonymous") : partner?.name || t("user");
	const partnerAge =
		!isHidden && partner?.profile?.age
			? t("yearsOld", { age: partner.profile.age })
			: null;
	const partnerGender =
		!isHidden &&
		partner?.profile?.is_male !== null &&
		partner?.profile?.is_male !== undefined
			? partner.profile.is_male
				? t("male")
				: t("female")
			: null;
	const partnerSubtitle = [partnerAge, partnerGender].filter(Boolean).join(" • ");

	if (isAdminOpen) {
		return (
			<header className="absolute left-0 right-0 top-0 flex h-16 shrink-0 items-center justify-between border-b-2 px-4">
				<div className="flex items-center gap-2">
					{trigger}
					<span className="text-base font-semibold">{t("adminPanel")}</span>
				</div>
				<button
					onClick={() => setIsAdminOpen(false)}
					className="flex h-8 w-8 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					title={t("closeAdminPanel")}
				>
					<X size={16} />
				</button>
			</header>
		);
	}

	return (
		<header className="absolute left-0 right-0 top-0 flex h-16 shrink-0 items-center justify-between border-b-2 px-4">
			<div className="flex items-center gap-2">
				{trigger}
				{partner && (
					<div className="flex flex-col leading-tight">
						<span className="text-base font-semibold">{partnerName}</span>
						{partnerSubtitle && (
							<span className="text-xs text-muted-foreground">
								{partnerSubtitle}
							</span>
						)}
					</div>
				)}
			</div>
			<div className="flex items-center gap-2">
				{partner && !reported && (
					<button
						onClick={handleReport}
						className="flex h-8 w-8 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-950"
						title={t("reportUser")}
					>
						<Flag size={16} />
					</button>
				)}
				<AppActionButton />
			</div>
		</header>
	);
}
