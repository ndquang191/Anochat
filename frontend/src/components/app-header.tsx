"use client";

import React from "react";
import { X } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage } from "@/contexts/theme";
import { AppActionButton } from "@/components/header/app-action-button";

interface HeaderProps {
	trigger: React.ReactNode;
}

export default function AppHeader({ trigger }: HeaderProps) {
	const { room } = useAuth();
	const { isAdminOpen, setIsAdminOpen } = useAdmin();
	const { t } = useLanguage();
	const partner = room?.partner;

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
				<AppActionButton />
			</div>
		</header>
	);
}
