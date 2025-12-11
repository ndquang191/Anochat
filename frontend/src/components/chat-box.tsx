"use client";

import * as React from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { useWebSocketChat } from "@/hooks/use-websocket-chat";
import { useAuth } from "@/contexts/auth";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

export default function ChatBox() {
	const { user } = useAuth();
	const [inputMessage, setInputMessage] = React.useState("");
	const messagesEndRef = React.useRef<HTMLDivElement>(null);
	const typingTimeoutRef = React.useRef<NodeJS.Timeout | null>(null);

	const { messages, sendMessage, isConnected, roomId, leaveRoom, isPartnerTyping, sendTypingIndicator } =
		useWebSocketChat({
			userId: user?.id || "",
			onMatchFound: (roomId) => {
				console.log("🔄 onMatchFound called with:", roomId);
				toast.success("Đã tìm thấy đối tác chat!", {
					description: `Bạn đã được kết nối với người dùng khác`,
				});
			},
			onPartnerLeft: () => {
				toast.info("Đối tác đã rời phòng", {
					description: "Bạn có thể tìm kiếm đối tác mới",
				});
			},
		});

	const scrollToBottom = React.useCallback(() => {
		if (messagesEndRef.current) {
			messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
		}
	}, []);

	React.useEffect(() => {
		scrollToBottom();
	}, [messages, scrollToBottom]);

	const handleSendMessage = React.useCallback(
		(e: React.FormEvent) => {
			e.preventDefault();
			if (inputMessage.trim() === "" || !roomId) return;

			sendMessage(inputMessage);
			setInputMessage("");

			// Stop typing indicator
			sendTypingIndicator(false);
		},
		[inputMessage, roomId, sendMessage, sendTypingIndicator]
	);

	const handleInputChange = React.useCallback(
		(e: React.ChangeEvent<HTMLInputElement>) => {
			setInputMessage(e.target.value);

			// Send typing indicator
			if (roomId && e.target.value.length > 0) {
				sendTypingIndicator(true);

				// Clear previous timeout
				if (typingTimeoutRef.current) {
					clearTimeout(typingTimeoutRef.current);
				}

				// Stop typing indicator after 2 seconds of inactivity
				typingTimeoutRef.current = setTimeout(() => {
					sendTypingIndicator(false);
				}, 2000);
			}
		},
		[roomId, sendTypingIndicator]
	);

	const handleLeaveRoom = React.useCallback(() => {
		if (roomId) {
			leaveRoom();
			toast.info("Bạn đã rời khỏi phòng chat");
		}
	}, [roomId, leaveRoom]);

	// Show loading if no user yet
	if (!user) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
					<p className="text-muted-foreground">Đang tải thông tin người dùng...</p>
				</div>
			</div>
		);
	}

	// Show connection status
	if (!isConnected) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
					<p className="text-muted-foreground">Đang kết nối WebSocket...</p>
				</div>
			</div>
		);
	}

	// Show "no room" state
	if (!roomId) {
		return (
			<div className="h-full w-full flex items-center justify-center">
				<div className="text-center space-y-4">
					<p className="text-lg text-muted-foreground">Chưa có phòng chat nào</p>
					<p className="text-sm text-muted-foreground">
						Vui lòng tham gia hàng chờ để tìm đối tác chat
					</p>
				</div>
			</div>
		);
	}

	return (
		<div className="flex flex-col bg-card text-card-foreground shadow-sm h-full relative">
			{/* Chat header */}
			<div className="border-b p-4 flex justify-between items-center">
				<div>
					<h3 className="font-semibold">Phòng chat</h3>
					<p className="text-sm text-muted-foreground">
						{isPartnerTyping ? "Đang nhập..." : "Đang trực tuyến"}
					</p>
				</div>
				<Button variant="outline" size="sm" onClick={handleLeaveRoom}>
					Rời phòng
				</Button>
			</div>

			{/* Messages area */}
			<ScrollArea className="flex-1 p-2 h-[calc(100vh-200px)]">
				<div className="flex flex-col gap-3">
					{messages.length === 0 ? (
						<div className="text-center text-muted-foreground py-8">
							<p>Chưa có tin nhắn nào</p>
							<p className="text-sm mt-2">Hãy bắt đầu cuộc trò chuyện!</p>
						</div>
					) : (
						messages.map((message) => {
							const isCurrentUser = message.sender_id === user?.id;
							return (
								<div
									key={message.id}
									className={`flex ${isCurrentUser ? "justify-end" : "justify-start"}`}
								>
									<div
										className={`max-w-[70%] rounded-lg p-3 group relative ${isCurrentUser
											? "bg-[#516b91] text-white"
											: "bg-gray-200 text-gray-900 dark:bg-gray-800 dark:text-gray-200"
											}`}
									>
										<p>{message.content}</p>
										{/* <p className="text-xs mt-1 opacity-0 group-hover:opacity-70 transition-opacity">
											{new Date(message.created_at * 1000).toLocaleTimeString("vi-VN", {
												hour: "2-digit",
												minute: "2-digit",
											})}
										</p> */}
									</div>
								</div>
							);
						})
					)}
					<div ref={messagesEndRef} />
				</div>
			</ScrollArea>

			{/* Input form */}
			<form
				onSubmit={handleSendMessage}
				className="absolute bottom-0 left-0 right-0 flex items-center gap-2 border-t p-4 bg-background"
			>
				<Input
					placeholder="Nhập tin nhắn của bạn..."
					value={inputMessage}
					onChange={handleInputChange}
					className="flex-1"
					disabled={!roomId}
				/>
				<Button type="submit" className="shrink-0" disabled={!roomId || !inputMessage.trim()}>
					Gửi
				</Button>
			</form>
		</div>
	);
}
