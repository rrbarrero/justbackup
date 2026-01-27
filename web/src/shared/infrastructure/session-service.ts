export class SessionService {
  private static readonly TOKEN_KEY = "token";

  public static setSession(token: string): void {
    if (typeof window !== "undefined") {
      localStorage.setItem(this.TOKEN_KEY, token);
      document.cookie = `${this.TOKEN_KEY}=${token}; path=/; max-age=86400; SameSite=Strict`;
    }
  }

  public static getToken(): string | null {
    if (typeof window !== "undefined") {
      return localStorage.getItem(this.TOKEN_KEY);
    }
    return null;
  }

  public static clearSession(): void {
    if (typeof window !== "undefined") {
      localStorage.removeItem(this.TOKEN_KEY);
      document.cookie = `${this.TOKEN_KEY}=; path=/; expires=Thu, 01 Jan 1970 00:00:01 GMT;`;
    }
  }

  public static logout(): void {
    this.clearSession();
    if (typeof window !== "undefined") {
      window.location.href = "/login";
    }
  }
}
