/**
 * Breakpoint reactif dùng chung cho lớp giao diện điện thoại (Cờ vua Dương Sinh).
 *
 * Vì sao dùng `window.matchMedia` chứ không `window.innerWidth`:
 * - `<html>` mang CSS `zoom` (xem `composables/useFont.ts` + `utils/zoom.ts`), nên
 *   `window.innerWidth` trả **visual pixel** đã nhân zoom → so sánh trực tiếp với
 *   một hằng số px sẽ lệch khi Thầy đổi cỡ chữ.
 * - Media query (cả trong CSS lẫn `matchMedia`) được đánh giá ở **tầng viewport**,
 *   KHÔNG bị `zoom` trên `<html>` nhân. Nên JS ở đây và CSS trong
 *   `assets/theme/duongsinh-responsive.css` luôn lật cùng một lúc, không cần chia zoom.
 *
 * Nếu cần ĐO/GHI toạ độ px cho popover thì vẫn phải dùng `cssViewportSize()` /
 * `rectToCssPx()` trong `utils/zoom.ts` — đó là bài toán khác.
 */
import { readonly, ref } from 'vue';

/**
 * Giữ ĐỒNG BỘ với `@media screen and (max-width: 767px)` trong
 * `assets/theme/duongsinh-responsive.css`. Có test khoá bất biến này:
 * `assets/theme/responsive.style.test.mjs`.
 *
 * Chọn 767 (không phải 600) vì khung tối thiểu trước đợt sửa là
 * sidebar 260px + vùng chat 400px = 660px — đặt 600 thì dải 601–767 vẫn vỡ.
 */
export const BP = { mobile: 767 } as const;

const MOBILE_QUERY = `screen and (max-width: ${BP.mobile}px)`;

const mq: MediaQueryList | null =
  typeof window !== 'undefined' && typeof window.matchMedia === 'function'
    ? window.matchMedia(MOBILE_QUERY)
    : null;

const mobileState = ref(mq?.matches ?? false);

if (mq) {
  const onChange = (e: MediaQueryListEvent) => {
    mobileState.value = e.matches;
  };
  // addEventListener có từ Safari 14; addListener là đường lùi cho WebView cũ.
  if (typeof mq.addEventListener === 'function') {
    mq.addEventListener('change', onChange);
  } else if (typeof (mq as any).addListener === 'function') {
    (mq as any).addListener(onChange);
  }
}

/** `true` khi viewport ≤ 767px (điện thoại / tablet dọc hẹp). */
export const isMobile = readonly(mobileState);

export function useBreakpoint() {
  return { isMobile, BP };
}
