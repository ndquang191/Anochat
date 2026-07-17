"use client";

import { useState, useCallback, useRef, useEffect, type FormEvent, type ChangeEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { convertEmotes } from "@/lib/emotes";
import { useLanguage } from "@/contexts/theme";
import { MAX_MESSAGE_LENGTH } from "@/types";

interface ChatInputProps {
	onSendMessage: (message: string) => void;
	onTypingChange?: (isTyping: boolean) => void;
	disabled: boolean;
}

const TYPING_TIMEOUT = 2000;

export function ChatInput({ onSendMessage, onTypingChange, disabled }: ChatInputProps) {
	const { t } = useLanguage();
	const [message, setMessage] = useState("");
	const inputRef = useRef<HTMLInputElement>(null);
	const typingTimeoutRef = useRef<number | null>(null);

	useEffect(() => {
		const handleFocus = () => inputRef.current?.focus();
		window.addEventListener("focus", handleFocus);
		return () => window.removeEventListener("focus", handleFocus);
	}, []);

	useEffect(() => {
		return () => {
			if (typingTimeoutRef.current !== null) {
				window.clearTimeout(typingTimeoutRef.current);
			}
			onTypingChange?.(false);
		};
	}, [onTypingChange]);

	const handleSubmit = useCallback(
		(e: FormEvent) => {
			e.preventDefault();
			if (message.trim() === "" || disabled) return;

			onSendMessage(convertEmotes(message));
			setMessage("");
			onTypingChange?.(false);
			if (typingTimeoutRef.current !== null) {
				window.clearTimeout(typingTimeoutRef.current);
				typingTimeoutRef.current = null;
			}
		},
		[message, disabled, onSendMessage, onTypingChange]
	);

	const handleChange = useCallback(
		(e: ChangeEvent<HTMLInputElement>) => {
			setMessage(e.target.value);
			if (!onTypingChange) return;

			onTypingChange(e.target.value.trim().length > 0);
			if (typingTimeoutRef.current !== null) {
				window.clearTimeout(typingTimeoutRef.current);
			}
			typingTimeoutRef.current = window.setTimeout(() => {
				onTypingChange(false);
				typingTimeoutRef.current = null;
			}, TYPING_TIMEOUT);
		},
		[onTypingChange]
	);

	return (
		<form
			onSubmit={handleSubmit}
			className="flex items-center gap-2 border-t bg-background p-4 shrink-0"
		>
			<Input
				ref={inputRef}
				placeholder={t("yourMessagePlaceholder")}
				value={message}
				onChange={handleChange}
				maxLength={MAX_MESSAGE_LENGTH}
				className="flex-1"
				disabled={disabled}
			/>
			<Button type="submit" className="h-9 shrink-0" disabled={disabled || !message.trim()}>
				{t("send")}
			</Button>
		</form>
	);
}
