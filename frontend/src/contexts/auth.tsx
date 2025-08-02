"use client";

import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { User, AuthState, AuthContextType } from "@/types";
import { setCookie, getCookie, deleteCookie } from "@/lib/cookies";

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
	const [authState, setAuthState] = useState<AuthState>({
		isAuthenticated: false,
		user: null,
		token: null,
		loading: true,
	});

	console.log("AuthProvider render:", authState);

	useEffect(() => {
		// Check for existing token on mount
		const token = getCookie("jwt_token");
		const userInfo = getCookie("user_info");

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
				logout().catch(console.error);
			}
		} else {
			setAuthState((prev) => ({ ...prev, loading: false }));
		}
	}, []);

	const login = (token: string, user: User) => {
		console.log("Auth login called:", {
			token: token ? token.substring(0, 20) + "..." : "from-cookie",
			user,
		});

		// If token is provided, set it in cookie (for backward compatibility)
		if (token) {
			setCookie("jwt_token", token, 7); // 7 days
		}
		// Token is already set in HTTP-only cookie by backend

		setCookie("user_info", JSON.stringify(user), 7);
		setAuthState({
			isAuthenticated: true,
			user,
			token: token || getCookie("jwt_token"), // Use provided token or get from cookie
			loading: false,
		});
		console.log("Auth state updated, isAuthenticated:", true);
	};

	const logout = async () => {
		console.log("Logout called, calling backend to clear cookies");

		try {
			// Call backend to clear HTTP-only cookie
			const response = await fetch("http://localhost:8080/auth/logout", {
				method: "POST",
				credentials: "include", // Include cookies
			});

			if (response.ok) {
				console.log("Backend logout successful");
			} else {
				console.error("Backend logout failed:", response.status);
			}
		} catch (error) {
			console.error("Error calling logout endpoint:", error);
		}

		// Clear frontend cookies
		deleteCookie("jwt_token");
		deleteCookie("user_info");
		console.log("Frontend cookies cleared");

		setAuthState({
			isAuthenticated: false,
			user: null,
			token: null,
			loading: false,
		});

		// Force redirect to login after logout
		window.location.href = "/login";
	};

	const checkAuth = async () => {
		const token = getCookie("jwt_token");
		if (!token) {
			await logout();
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
				await logout();
			}
		} catch (error) {
			console.error("Auth check error:", error);
			await logout();
		}
	};

	const value: AuthContextType = {
		...authState,
		login,
		logout,
		checkAuth,
	};

	return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
	const context = useContext(AuthContext);
	if (!context) {
		throw new Error("useAuth must be used within AuthProvider");
	}
	return context;
}
