"use client";

import React from "react";
import { useAuth } from "@/contexts/auth";
import { ActionButton } from "./header/action-button";

interface HeaderProps {
	trigger: React.ReactNode;
}

export default function Header({ trigger }: HeaderProps) {
	const { room } = useAuth();
	const partner = room?.partner;
	const isHidden = partner?.profile?.is_hidden;

	const partnerName = isHidden ? "Ẩn danh" : (partner?.name || "Người dùng");
	const partnerAge = !isHidden && partner?.profile?.age ? `${partner.profile.age} tuổi` : null;
	const partnerGender =
		!isHidden && partner?.profile?.is_male !== null && partner?.profile?.is_male !== undefined
			? partner.profile.is_male ? "Nam" : "Nữ"
			: null;
	const partnerSubtitle = [partnerAge, partnerGender].filter(Boolean).join(" • ");

	return (
		<header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b-2 px-4">
			<div className="flex items-center gap-2">
				{trigger}
				{partner ? (
					<div className="flex flex-col leading-tight">
						<span className="text-base font-semibold">{partnerName}</span>
						{partnerSubtitle && (
							<span className="text-xs text-muted-foreground">{partnerSubtitle}</span>
						)}
					</div>
				) : (
					<h1 className="text-xl font-semibold">Chat ẩn danh</h1>
				)}
			</div>
			<div className="flex items-center gap-2">
				<ActionButton />
			</div>
		</header>
	);
}
