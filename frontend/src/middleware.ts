import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
	const token = request.cookies.get("jwt_token")?.value;
	const pathname = request.nextUrl.pathname;

	const publicPaths = ["/login", "/callback", "/error"];
	const isPublicPath = publicPaths.some((path) => pathname.startsWith(path));

	if (token && pathname === "/login") {
		return NextResponse.redirect(new URL("/", request.url));
	}

	if (isPublicPath) {
		return NextResponse.next();
	}

	if (!token) {
		const loginUrl = new URL("/login", request.url);
		loginUrl.searchParams.set("redirect", pathname);
		return NextResponse.redirect(loginUrl);
	}

	return NextResponse.next();
}

export const config = {
	matcher: [
		"/((?!api|_next/static|_next/image|favicon.ico).*)",
	],
};
