"use client";

import React from "react";
import { Plus, X } from "lucide-react";
import { useAuth } from "@/contexts/auth";
import { useAdmin } from "@/contexts/admin";
import { useLanguage } from "@/contexts/theme";
import { AppActionButton } from "@/components/header/app-action-button";
import { useQueue } from "@/hooks/use-queue";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface HeaderProps {
	trigger: React.ReactNode;
}

export default function AppHeader({ trigger }: HeaderProps) {
	const {
		room,
		rooms,
		maxActiveRooms,
		inQueue,
		selectRoom,
		unreadByRoom,
	} = useAuth();
	const { joinQueue, leaveQueue, isLoading: isQueueLoading } = useQueue();
	const { isAdminOpen, setIsAdminOpen } = useAdmin();
	const { language, t } = useLanguage();
	const atRoomLimit = rooms.length >= maxActiveRooms;
	const queueButtonLabel = atRoomLimit
		? language === "vi"
			? `Tối đa ${maxActiveRooms} phòng`
			: `Maximum ${maxActiveRooms} rooms`
		: inQueue
			? language === "vi"
				? "Đang tìm — bấm để huỷ"
				: "Searching — click to cancel"
			: language === "vi"
				? "Tìm thêm phòng"
				: "Find another room";

	React.useEffect(() => {
		if (isAdminOpen || rooms.length < 2) return;

		const handleArrowNavigation = (event: KeyboardEvent) => {
			if (event.key !== "ArrowLeft" && event.key !== "ArrowRight") return;
			if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
			const target = event.target;
			if (
				target instanceof HTMLElement &&
				target.closest('input, textarea, select, [contenteditable="true"], [role="textbox"]')
			) {
				return;
			}

			const currentIndex = rooms.findIndex((item) => item.id === room?.id);
			const direction = event.key === "ArrowRight" ? 1 : -1;
			const nextIndex = (Math.max(currentIndex, 0) + direction + rooms.length) % rooms.length;
			event.preventDefault();
			selectRoom(rooms[nextIndex].id);
		};

		window.addEventListener("keydown", handleArrowNavigation);
		return () => window.removeEventListener("keydown", handleArrowNavigation);
	}, [isAdminOpen, room?.id, rooms, selectRoom]);

	if (isAdminOpen) {
		return (
			<header className="absolute left-0 right-0 top-0 flex h-12 shrink-0 items-center justify-between border-b-2 px-4">
				<div className="flex items-center gap-1.5">
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
		<header className="absolute left-0 right-0 top-0 flex h-12 shrink-0 items-end gap-3 bg-background px-3 md:px-4">
			<span
				aria-hidden="true"
				className="pointer-events-none absolute inset-x-0 bottom-0 z-0 h-px bg-border"
			/>
			<div className="mb-1 flex h-9 shrink-0 items-center">{trigger}</div>
			{rooms.length > 0 && (
				<div className="flex h-full min-w-0 flex-1 touch-pan-x snap-x snap-proximity items-end gap-1.5 overflow-x-auto overscroll-x-contain scroll-smooth pt-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
					{rooms.map((item, index) => {
						const active = item.id === room?.id;
						const unread = unreadByRoom[item.id] ?? 0;
						const hidden = item.partner?.profile?.is_hidden;
						const label = hidden
							? t("anonymous")
							: item.partner?.nickname?.trim() || item.partner?.name || `${t("user")} ${index + 1}`;
						const age = !hidden ? item.partner?.profile?.age : undefined;
						return (
							<button
								key={item.id}
								type="button"
								onClick={() => selectRoom(item.id)}
								className={cn(
									"relative flex h-10 w-36 shrink-0 snap-start items-center justify-between gap-2 border border-transparent bg-transparent px-3 text-sm font-medium text-foreground sm:w-40 sm:gap-3 sm:px-4 md:w-48",
									active
										? "z-10 rounded-t-lg rounded-b-none border-primary border-b-transparent bg-card"
										: "before:absolute before:inset-x-0 before:inset-y-1 before:rounded-md before:border before:border-border before:bg-card before:transition-colors hover:before:border-primary/50",
								)}
								aria-current={active ? "page" : undefined}
								title={label}
							>
								<span className="relative z-10 truncate">{label}</span>
								{active && age != null && (
									<span className="relative z-10 shrink-0 rounded-md border border-primary/25 bg-primary/10 px-1.5 py-0.5 text-xs font-semibold tabular-nums text-primary">
										{age}
									</span>
								)}
								{!active && unread > 0 && (
									<span className="relative z-10 flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-unread px-1.5 text-[10px] font-bold leading-none tabular-nums text-unread-foreground ring-2 ring-card">
										{unread > 99 ? "99+" : unread}
									</span>
								)}
							</button>
						);
					})}
					<span className="mb-1.5 shrink-0" title={queueButtonLabel}>
						<Button
							type="button"
							size="icon"
							variant="outline"
							className={cn(
								"size-7 p-0",
								(isQueueLoading || inQueue) && !atRoomLimit && "queue-border-spinner",
							)}
							disabled={isQueueLoading || atRoomLimit}
							onClick={() => void (inQueue ? leaveQueue() : joinQueue())}
							aria-label={queueButtonLabel}
						>
							<span className="relative flex size-5 items-center justify-center">
								{atRoomLimit || inQueue ? (
									<X className="size-4" strokeWidth={1.75} />
								) : !isQueueLoading ? (
									<Plus className="size-4" strokeWidth={1.75} />
								) : null}
							</span>
						</Button>
					</span>
				</div>
			)}
			<div className="mb-1 ml-auto flex h-9 shrink-0 items-center gap-2">
				<AppActionButton />
			</div>
		</header>
	);
}
