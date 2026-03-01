"use client";

import React from "react";
import { useAuth } from "@/contexts/auth";
import { ActionButton } from "./header/action-button";
import { UserInfoBar } from "./header/user-info-bar";

interface HeaderProps {
	trigger: React.ReactNode;
}

export default function Header({ trigger }: HeaderProps) {
	const { room } = useAuth();

	return (
		<div>
			<header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b px-4">
				<div className="flex items-center gap-2">
					{trigger}
					<h1 className="text-xl font-semibold">Chat ẩn danh</h1>
				</div>
				<div className="flex items-center gap-2">
					<ActionButton />
				</div>
			</header>
			{room?.partner && <UserInfoBar partner={room.partner} />}
		</div>
	);
}
