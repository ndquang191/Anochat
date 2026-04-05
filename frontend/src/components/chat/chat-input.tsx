"use client";

import { useState, useCallback, useRef, useEffect, FormEvent, ChangeEvent } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { convertEmotes } from "@/lib/emotes";

interface ChatInputProps {
	onSendMessage: (message: string) => void;
	disabled: boolean;
}

export function ChatInput({ onSendMessage, disabled }: ChatInputProps) {
	const [message, setMessage] = useState("");
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		const handleFocus = () => inputRef.current?.focus();
		window.addEventListener("focus", handleFocus);
		return () => window.removeEventListener("focus", handleFocus);
	}, []);

	const handleSubmit = useCallback(
		(e: FormEvent) => {
			e.preventDefault();
			if (message.trim() === "" || disabled) return;

			onSendMessage(convertEmotes(message));
			setMessage("");
		},
		[message, disabled, onSendMessage]
	);

	const handleChange = useCallback((e: ChangeEvent<HTMLInputElement>) => {
		setMessage(e.target.value);
	}, []);

	return (
		<form
			onSubmit={handleSubmit}
			className="flex items-center gap-2 border-t p-4 bg-background shrink-0"
		>
			<Input
				ref={inputRef}
				placeholder="Nhập tin nhắn của bạn..."
				value={message}
				onChange={handleChange}
				className="flex-1"
				disabled={disabled}
			/>
			<Button type="submit" className="shrink-0 h-9" disabled={disabled || !message.trim()}>
				Gửi
			</Button>
		</form>
	);
}
