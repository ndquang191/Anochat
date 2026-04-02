interface ChatMessageProps {
	content: string;
	isCurrentUser: boolean;
}

export function ChatMessage({ content, isCurrentUser }: ChatMessageProps) {
	return (
		<div className={`flex ${isCurrentUser ? "justify-end" : "justify-start"}`}>
			<div
				className={`max-w-[70%] rounded-md px-3 py-2 ${
					isCurrentUser
						? "bg-primary text-primary-foreground"
						: "bg-muted text-foreground"
				}`}
			>
				<p>{content}</p>
			</div>
		</div>
	);
}
