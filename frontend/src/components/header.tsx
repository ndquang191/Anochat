"use client";
import React from "react";
import { Loader2, RotateCw, X } from "lucide-react";

function ConnectDisconnectButton() {
	const [isConnected, setIsConnected] = React.useState(false);
	const [isConnecting, setIsConnecting] = React.useState(false);

	const handleClick = () => {
		if (isConnected) {
			setIsConnected(false);
		} else if (!isConnecting) {
			setIsConnecting(true);
			// Simulate connection
			setTimeout(() => {
				setIsConnecting(false);
				setIsConnected(true);
			}, 2000);
		}
	};

	const iconBaseClass = `absolute inset-0 flex items-center justify-center text-white transition-all duration-300`;
	const getIconClass = (visible: boolean, spinning = false) =>
		`${iconBaseClass} ${visible ? "opacity-100 scale-100 rotate-0" : "opacity-0 scale-0 rotate-90"} ${
			spinning ? "animate-spin" : ""
		}`;

	return (
		<button
			onClick={handleClick}
			disabled={isConnecting}
			className={`relative w-10 h-10 rounded-full transition-all duration-300 ease-in-out transform hover:scale-110 bg-primary hover:bg-primary/90`}
		>
			{/* Icon loading tìm người */}
			<div className={getIconClass(isConnecting, true)}>
				<Loader2 size={18} />
			</div>

			{/* Icon đang trò chuyện */}
			<div className={getIconClass(isConnected)}>
				<X size={18} />
			</div>

			{/* Icon ban đầu (chưa kết nối) */}
			<div className={getIconClass(!isConnected && !isConnecting)}>
				<RotateCw size={18} />
			</div>
		</button>
	);
}

const Header = ({ trigger }: { trigger: React.ReactNode }) => {
	return (
		<div>
			<header className="absolute top-0 left-0 right-0 flex h-16 shrink-0 items-center justify-between border-b px-4">
				<div className="flex items-center gap-2">
					{trigger}
					<h1 className="text-xl font-semibold">Chat ẩn danh</h1>
				</div>
				<div className="flex items-center">
					<ConnectDisconnectButton />
				</div>
			</header>
		</div>
	);
};

export default Header;
