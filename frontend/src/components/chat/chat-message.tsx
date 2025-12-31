interface ChatMessageProps {
	content: string;
	isCurrentUser: boolean;
}

export function ChatMessage({ content, isCurrentUser }: ChatMessageProps) {
	return (
		<div className={`flex ${isCurrentUser ? "justify-end" : "justify-start"}`}>
			<div
				className={`max-w-[70%] rounded-lg p-3 ${
					isCurrentUser
						? "bg-[#516b91] text-white"
						: "bg-gray-200 text-gray-900 dark:bg-gray-800 dark:text-gray-200"
				}`}
			>
				<p>{content}</p>
			</div>
		</div>
	);
}
