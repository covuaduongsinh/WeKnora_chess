import { get, post, put, del } from "../../utils/request";

// API quản lý khóa học & bài học cờ vua (Phase 4 LMS).
// Lưu ý: path phải bao gồm tiền tố /api/v1 (giống các api khác — baseURL không chứa nó).

export interface ChessCourse {
  id: string;
  title: string;
  description: string;
  level: string;
  cover_url: string;
  sort_order: number;
  lesson_count?: number;
  slug?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ChessLesson {
  id: string;
  course_id: string;
  title: string;
  content: string;
  fen: string;
  pgn: string;
  sort_order: number;
  slug?: string;
  created_at?: string;
  updated_at?: string;
}

// ---- Khóa học ----
export const listCourses = () => get("/api/v1/chess/courses");
export const getCourse = (id: string) => get(`/api/v1/chess/courses/${id}`);
// Giải mã wikilink [[course/<slug>]] → khóa học.
export const getCourseBySlug = (slug: string) => get(`/api/v1/chess/courses/by-slug/${encodeURIComponent(slug)}`);
export const getCourseBacklinks = (slug: string) => get(`/api/v1/chess/courses/by-slug/${encodeURIComponent(slug)}/backlinks`);
export const createCourse = (data: Partial<ChessCourse>) => post("/api/v1/chess/courses", data);
export const updateCourse = (id: string, data: Partial<ChessCourse>) => put(`/api/v1/chess/courses/${id}`, data);
export const deleteCourse = (id: string) => del(`/api/v1/chess/courses/${id}`);
// Export/Import khóa học (kèm bài học) dạng JSON — sao lưu/chia sẻ. Import luôn tạo mới.
export const exportCourses = () => get("/api/v1/chess/courses/export");
export const importCourses = (courses: any[]) => post("/api/v1/chess/courses/import", { courses });

// ---- Bài học ----
export const listLessons = (courseId: string) => get(`/api/v1/chess/courses/${courseId}/lessons`);
export const createLesson = (courseId: string, data: Partial<ChessLesson>) =>
  post(`/api/v1/chess/courses/${courseId}/lessons`, data);
export const getLesson = (lessonId: string) => get(`/api/v1/chess/lessons/${lessonId}`);
// Giải mã wikilink [[lesson/<slug>]] → bài giảng.
export const getLessonBySlug = (slug: string) => get(`/api/v1/chess/lessons/by-slug/${encodeURIComponent(slug)}`);
export const getLessonBacklinks = (slug: string) => get(`/api/v1/chess/lessons/by-slug/${encodeURIComponent(slug)}/backlinks`);
export const updateLesson = (lessonId: string, data: Partial<ChessLesson>) =>
  put(`/api/v1/chess/lessons/${lessonId}`, data);
export const deleteLesson = (lessonId: string) => del(`/api/v1/chess/lessons/${lessonId}`);

// ---- Kho ván đấu ----
export interface ChessGame {
  id: string;
  white: string; black: string; result: string;
  eco: string; event: string; date: string;
  pgn: string; ply_count: number;
  slug?: string;
  created_at?: string;
}
function qs(params: Record<string, string>): string {
  const p = Object.entries(params).filter(([, v]) => v).map(([k, v]) => `${k}=${encodeURIComponent(v)}`);
  return p.length ? `?${p.join("&")}` : "";
}
export const listGames = (f: Partial<{ white: string; black: string; eco: string; result: string }> = {}) =>
  get(`/api/v1/chess/games${qs(f as Record<string, string>)}`);
export const getGame = (id: string) => get(`/api/v1/chess/games/${id}`);
// Giải mã wikilink [[game/<slug>]] → ván cờ.
export const getGameBySlug = (slug: string) => get(`/api/v1/chess/games/by-slug/${encodeURIComponent(slug)}`);
export const getGameBacklinks = (slug: string) => get(`/api/v1/chess/games/by-slug/${encodeURIComponent(slug)}/backlinks`);
export const createGame = (data: Partial<ChessGame>) => post("/api/v1/chess/games", data);
export const updateGame = (id: string, data: Partial<ChessGame>) => put(`/api/v1/chess/games/${id}`, data);
export const deleteGame = (id: string) => del(`/api/v1/chess/games/${id}`);
export const importGames = (pgn: string) => post("/api/v1/chess/games/import", { pgn });
// Export ván đấu (theo bộ lọc) thành PGN nhiều ván.
export const exportGamesPGN = (f: Partial<{ white: string; black: string; eco: string; result: string }> = {}) =>
  get(`/api/v1/chess/games/export${qs(f as Record<string, string>)}`);

// ---- Ngân hàng bài tập ----
export interface ChessPuzzle {
  id: string;
  title: string; fen: string; solution: string;
  theme: string; difficulty: string; source: string;
  slug?: string;
  created_at?: string;
}
export const listPuzzles = (f: Partial<{ theme: string; difficulty: string }> = {}) =>
  get(`/api/v1/chess/puzzles${qs(f as Record<string, string>)}`);
export const getPuzzle = (id: string) => get(`/api/v1/chess/puzzles/${id}`);
// Giải mã wikilink [[puzzle/<slug>]] → thế cờ/bài tập.
export const getPuzzleBySlug = (slug: string) => get(`/api/v1/chess/puzzles/by-slug/${encodeURIComponent(slug)}`);
export const getPuzzleBacklinks = (slug: string) => get(`/api/v1/chess/puzzles/by-slug/${encodeURIComponent(slug)}/backlinks`);
export const createPuzzle = (data: Partial<ChessPuzzle>) => post("/api/v1/chess/puzzles", data);
export const updatePuzzle = (id: string, data: Partial<ChessPuzzle>) => put(`/api/v1/chess/puzzles/${id}`, data);
export const deletePuzzle = (id: string) => del(`/api/v1/chess/puzzles/${id}`);
export const randomPuzzle = (f: Partial<{ theme: string; difficulty: string }> = {}) =>
  get(`/api/v1/chess/puzzles/random${qs(f as Record<string, string>)}`);
// Export/Import bài tập dạng JSON — sao lưu/chia sẻ. Import luôn tạo mới.
export const exportPuzzles = (f: Partial<{ theme: string; difficulty: string }> = {}) =>
  get(`/api/v1/chess/puzzles/export${qs(f as Record<string, string>)}`);
export const importPuzzles = (puzzles: any[]) => post("/api/v1/chess/puzzles/import", { puzzles });
