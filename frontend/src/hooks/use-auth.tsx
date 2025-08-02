"use client";

import { useState, useEffect } from "react";

interface User {
	id: string;
	email: string;
	name: string;
	avatar_url: string;
}

interface AuthState {
	isAuthenticated: boolean;
	user: User | null;
	token: string | null;
	loading: boolean;
}

export function useAuth() {
	const [authState, setAuthState] = useState<AuthState>({
		isAuthenticated: false,
		user: null,
		token: null,
		loading: true,
	});

	useEffect(() => {
		// Check for existing token on mount
		const token = localStorage.getItem("jwt_token");
		const userInfo = localStorage.getItem("user_info");

		if (token && userInfo) {
			try {
				const user = JSON.parse(userInfo);
				setAuthState({
					isAuthenticated: true,
					user,
					token,
					loading: false,
				});
			} catch (error) {
				console.error("Error parsing user info:", error);
				logout();
			}
		} else {
			setAuthState((prev) => ({ ...prev, loading: false }));
		}
	}, []);

	const login = (token: string, user: User) => {
		localStorage.setItem("jwt_token", token);
		localStorage.setItem("user_info", JSON.stringify(user));
		setAuthState({
			isAuthenticated: true,
			user,
			token,
			loading: false,
		});
	};

	const logout = () => {
		localStorage.removeItem("jwt_token");
		localStorage.removeItem("user_info");
		setAuthState({
			isAuthenticated: false,
			user: null,
			token: null,
			loading: false,
		});
	};

	const checkAuth = async () => {
		const token = localStorage.getItem("jwt_token");
		if (!token) {
			logout();
			return;
		}

		try {
			// Call backend to validate token and get user state
			const response = await fetch("http://localhost:8080/user/state", {
				headers: {
					Authorization: `Bearer ${token}`,
				},
			});

			if (response.ok) {
				const data = await response.json();
				setAuthState({
					isAuthenticated: true,
					user: data.user,
					token,
					loading: false,
				});
			} else {
				logout();
			}
		} catch (error) {
			console.error("Auth check error:", error);
			logout();
		}
	};

	return {
		...authState,
		login,
		logout,
		checkAuth,
	};
}
