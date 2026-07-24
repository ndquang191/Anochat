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
	const partnerName = isHidden
		? t("anonymous")
		: partner?.nickname?.trim() || partner?.name || t("user");
	const partnerAge =
		!isHidden && partner?.profile?.age ? partner.profile.age : null;
	const ageBadgeClass =
		partner?.profile?.is_male === true
			? "border-blue-500/25 bg-blue-500/10 text-blue-700 dark:text-blue-300"
			: partner?.profile?.is_male === false
				? "border-pink-500/25 bg-pink-500/10 text-pink-700 dark:text-pink-300"
				: "border-border bg-muted text-muted-foreground";

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
					<div className="flex min-w-0 items-center gap-2">
						<span className="truncate text-base font-semibold">{partnerName}</span>
						{partnerAge && (
							<span
								className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full border text-xs font-semibold tabular-nums ${ageBadgeClass}`}
								aria-label={t("yearsOld", { age: partnerAge })}
								title={t("yearsOld", { age: partnerAge })}
							>
								{partnerAge}
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
