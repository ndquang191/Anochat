"use client";

import { useRef, useEffect, useCallback, useMemo } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { ChatMessage } from "./chat-message";

interface Message {
	id: string;
	content: string;
	sender_id: string;
	created_at?: number;
}

interface ChatMessagesProps {
	messages: Message[];
	currentUserId: string;
}

type Row =
	| { kind: "date"; label: string }
	| { kind: "message"; message: Message };

function formatDateLabel(unixSec: number): string {
	const d = new Date(unixSec * 1000);
	const today = new Date();
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);
	if (d.toDateString() === today.toDateString()) return "Hôm nay";
	if (d.toDateString() === yesterday.toDateString()) return "Hôm qua";
	return d.toLocaleDateString("vi-VN", { day: "2-digit", month: "2-digit", year: "numeric" });
}

function buildRows(messages: Message[]): Row[] {
	const dateKeys = new Set(
		messages
			.filter((m) => m.created_at)
			.map((m) => new Date(m.created_at! * 1000).toDateString())
	);
	const multiDay = dateKeys.size > 1;

	const rows: Row[] = [];
	let lastDateKey = "";
	for (const message of messages) {
		const dateKey = message.created_at
			? new Date(message.created_at * 1000).toDateString()
			: "";
		if (multiDay && dateKey && dateKey !== lastDateKey) {
			rows.push({ kind: "date", label: formatDateLabel(message.created_at!) });
			lastDateKey = dateKey;
		}
		rows.push({ kind: "message", message });
	}
	return rows;
}

const SCROLL_THRESHOLD = 150;

export function ChatMessages({ messages, currentUserId }: ChatMessagesProps) {
	const scrollContainerRef = useRef<HTMLDivElement>(null);
	const isAtBottomRef = useRef(true);
	const prevMessageCountRef = useRef(0);

	const rows = useMemo(() => buildRows(messages), [messages]);

	const virtualizer = useVirtualizer({
		count: rows.length,
		getScrollElement: () => scrollContainerRef.current,
		estimateSize: (index) => (rows[index].kind === "date" ? 32 : 60),
		overscan: 5,
	});

	const handleScroll = useCallback(() => {
		const el = scrollContainerRef.current;
		if (!el) return;
		const distanceFromBottom = el.scrollHeight - (el.scrollTop + el.clientHeight);
		isAtBottomRef.current = distanceFromBottom < SCROLL_THRESHOLD;
	}, []);

	useEffect(() => {
		const prevCount = prevMessageCountRef.current;
		const nextCount = messages.length;
		prevMessageCountRef.current = nextCount;

		if (nextCount === 0) {
			isAtBottomRef.current = true;
			virtualizer.measure();
			return;
		}

		if (nextCount > prevCount && isAtBottomRef.current) {
			requestAnimationFrame(() => {
				const el = scrollContainerRef.current;
				if (el) el.scrollTop = el.scrollHeight;
			});
		}
	}, [messages, virtualizer]);

	if (messages.length === 0) {
		return (
			<div className="flex-1 min-h-0 flex items-center justify-center">
				<div className="text-center text-muted-foreground py-8">
					<p className="text-sm md:text-base">Chưa có tin nhắn nào</p>
					<p className="text-sm md:text-base mt-2">Hãy bắt đầu cuộc trò chuyện!</p>
				</div>
			</div>
		);
	}

	const virtualItems = virtualizer.getVirtualItems();

	return (
		<div
			ref={scrollContainerRef}
			onScroll={handleScroll}
			className="flex-1 min-h-0 overflow-y-auto overscroll-none px-2 pt-2"
		>
				<div
					style={{
						height: virtualizer.getTotalSize(),
						width: "100%",
						position: "relative",
					}}
				>
					<div
						style={{
							position: "absolute",
							top: 0,
							left: 0,
							width: "100%",
							transform: `translateY(${virtualItems[0]?.start ?? 0}px)`,
						}}
					>
						{virtualItems.map((virtualItem) => {
							const row = rows[virtualItem.index];
							return (
								<div
									key={virtualItem.key}
									data-index={virtualItem.index}
									ref={virtualizer.measureElement}
									className="pb-3"
								>
									{row.kind === "date" ? (
										<div className="flex items-center gap-3 py-1">
											<div className="flex-1 h-px bg-border" />
											<span className="text-xs text-muted-foreground">{row.label}</span>
											<div className="flex-1 h-px bg-border" />
										</div>
									) : (
										<ChatMessage
											content={row.message.content}
											isCurrentUser={row.message.sender_id === currentUserId}
											created_at={row.message.created_at}
										/>
									)}
								</div>
							);
						})}
					</div>
				</div>
			</div>
	);
}
