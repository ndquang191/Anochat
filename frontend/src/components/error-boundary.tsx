"use client";

import React from "react";
import { translateStored } from "@/lib/i18n";

interface Props {
	children: React.ReactNode;
	fallback?: React.ReactNode;
}

interface State {
	hasError: boolean;
	error: Error | null;
}

export class ErrorBoundary extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = { hasError: false, error: null };
	}

	static getDerivedStateFromError(error: Error): State {
		return { hasError: true, error };
	}

	componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
		console.error("ErrorBoundary caught:", error, errorInfo);
	}

	private handleSignOut = () => {
		const past = "Thu, 01 Jan 1970 00:00:00 UTC";
		document.cookie = `user_info=;expires=${past};path=/;`;
		document.cookie = `has_session=;expires=${past};path=/;`;
		window.location.href = "/login";
	};

	render() {
		if (this.state.hasError) {
			if (this.props.fallback) {
				return this.props.fallback;
			}

			return (
				<div className="flex min-h-svh w-full items-center justify-center p-6">
					<div className="text-center max-w-md">
						<h1 className="text-2xl font-bold mb-4">
							{translateStored("somethingWentWrong")}
						</h1>
						<p className="text-muted-foreground mb-4">
							{translateStored("comeBackLaterOrSignOut")}
						</p>
						<button
							onClick={this.handleSignOut}
							className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90 cursor-pointer"
						>
							{translateStored("signOut")}
						</button>
					</div>
				</div>
			);
		}

		return this.props.children;
	}
}
