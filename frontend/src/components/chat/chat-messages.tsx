"use client";

import { useRef, useEffect, useLayoutEffect, useCallback, useMemo } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { Loader2 } from "lucide-react";
import { ChatMessage } from "./chat-message";
import { useLanguage } from "@/contexts/theme";
import type { Language } from "@/lib/i18n";

interface Message {
	id: string;
	content: string;
	sender_id: string;
	created_at?: number;
}

interface ChatMessagesProps {
	messages: Message[];
	currentUserId: string;
	hasMore?: boolean;
	isLoadingOlder?: boolean;
	loadOlderError?: boolean;
	onLoadOlder?: () => Promise<boolean>;
}

type Row =
	| { kind: "date"; label: string }
	| { kind: "message"; message: Message };

function formatDateLabel(
	unixSec: number,
	language: Language,
	todayLabel: string,
	yesterdayLabel: string
): string {
	const d = new Date(unixSec * 1000);
	const today = new Date();
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);
	if (d.toDateString() === today.toDateString()) return todayLabel;
	if (d.toDateString() === yesterday.toDateString()) return yesterdayLabel;
	return d.toLocaleDateString(language === "vi" ? "vi-VN" : "en-US", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
	});
}

function buildRows(
	messages: Message[],
	language: Language,
	todayLabel: string,
	yesterdayLabel: string
): Row[] {
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
			rows.push({
				kind: "date",
				label: formatDateLabel(
					message.created_at!,
					language,
					todayLabel,
					yesterdayLabel
				),
			});
			lastDateKey = dateKey;
		}
		rows.push({ kind: "message", message });
	}
	return rows;
}

const SCROLL_THRESHOLD = 150;
const LOAD_OLDER_THRESHOLD = 80;

export function ChatMessages({
	messages,
	currentUserId,
	hasMore = false,
	isLoadingOlder = false,
	loadOlderError = false,
	onLoadOlder,
}: ChatMessagesProps) {
	const { language, t } = useLanguage();
	const scrollContainerRef = useRef<HTMLDivElement>(null);
	const isAtBottomRef = useRef(true);
	const prevMessageCountRef = useRef(0);
	const prependAnchorRef = useRef<{
		firstMessageId: string;
		scrollHeight: number;
		scrollTop: number;
	} | null>(null);
	const loadOlderInFlightRef = useRef(false);

	const todayLabel = t("today");
	const yesterdayLabel = t("yesterday");
	const rows = useMemo(
		() => buildRows(messages, language, todayLabel, yesterdayLabel),
		[messages, language, todayLabel, yesterdayLabel]
	);

	const virtualizer = useVirtualizer({
		count: rows.length,
		getScrollElement: () => scrollContainerRef.current,
		estimateSize: (index) => (rows[index].kind === "date" ? 28 : 44),
		overscan: 5,
	});

	const loadOlder = useCallback(async () => {
		const el = scrollContainerRef.current;
		if (
			!el ||
			!onLoadOlder ||
			!hasMore ||
			isLoadingOlder ||
			loadOlderInFlightRef.current
		) {
			return;
		}

		loadOlderInFlightRef.current = true;
		prependAnchorRef.current = {
			firstMessageId: messages[0]?.id ?? "",
			scrollHeight: el.scrollHeight,
			scrollTop: el.scrollTop,
		};
		try {
			const loaded = await onLoadOlder();
			if (!loaded) {
				prependAnchorRef.current = null;
			}
		} finally {
			loadOlderInFlightRef.current = false;
		}
	}, [hasMore, isLoadingOlder, messages, onLoadOlder]);

	const handleScroll = useCallback(() => {
		const el = scrollContainerRef.current;
		if (!el) return;
		const distanceFromBottom = el.scrollHeight - (el.scrollTop + el.clientHeight);
		isAtBottomRef.current = distanceFromBottom < SCROLL_THRESHOLD;
		if (el.scrollTop <= LOAD_OLDER_THRESHOLD) {
			void loadOlder();
		}
	}, [loadOlder]);

	useLayoutEffect(() => {
		const anchor = prependAnchorRef.current;
		const el = scrollContainerRef.current;
		if (!anchor || !el || messages[0]?.id === anchor.firstMessageId) return;

		virtualizer.measure();
		requestAnimationFrame(() => {
			const current = scrollContainerRef.current;
			if (!current) return;
			current.scrollTop =
				anchor.scrollTop + (current.scrollHeight - anchor.scrollHeight);
			prependAnchorRef.current = null;
		});
	}, [messages, virtualizer]);

	useEffect(() => {
		const prevCount = prevMessageCountRef.current;
		const nextCount = messages.length;
		prevMessageCountRef.current = nextCount;

		if (nextCount === 0) {
			isAtBottomRef.current = true;
			virtualizer.measure();
			return;
		}

		if (
			nextCount > prevCount &&
			isAtBottomRef.current &&
			!prependAnchorRef.current
		) {
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
					<p className="text-sm md:text-base">{t("noMessages")}</p>
					<p className="text-sm md:text-base mt-2">{t("startConversation")}</p>
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
				{(isLoadingOlder || loadOlderError) && (
					<div className="pointer-events-none sticky top-2 z-10 flex h-0 justify-center">
						{isLoadingOlder ? (
							<span className="rounded-full border bg-background/95 p-2 shadow-sm">
								<Loader2 className="size-4 animate-spin text-muted-foreground" />
							</span>
						) : (
							<button
								type="button"
								onClick={() => void loadOlder()}
								className="pointer-events-auto rounded-full border bg-background/95 px-3 py-1 text-xs text-muted-foreground shadow-sm hover:text-foreground"
							>
								{t("pleaseTryAgain")}
							</button>
						)}
					</div>
				)}
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
							const nextRow = rows[virtualItem.index + 1];
							const endsSenderGroup =
								row.kind === "message" &&
								(nextRow?.kind !== "message" ||
									nextRow.message.sender_id !== row.message.sender_id);
							return (
								<div
									key={virtualItem.key}
									data-index={virtualItem.index}
									ref={virtualizer.measureElement}
									className={
										row.kind === "date"
											? "pb-2"
											: endsSenderGroup
												? "pb-1.5"
												: "pb-1"
									}
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
