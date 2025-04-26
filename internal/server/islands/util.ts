export function getCookie(name: string) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return (parts.pop() ?? "").split(";").shift();
  return null;
}

export const myUserId = getCookie("user_id") ?? "";

export function stringToColor(str: string) {
  let hash = 0;

  // Generate hash
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i);
    hash = hash & hash; // Convert to 32bit integer
  }

  // Create separate hash values for R, G, B
  const r = (hash & 0xFF0000) >> 16;
  const g = (hash & 0x00FF00) >> 8;
  const b = hash & 0x0000FF;

  // Scale to mid-range (60-200) to avoid too dark or too bright
  const scaleToMidRange = (value: number) => {
    return Math.floor(60 + (value % 141));
  };

  // Format as hex
  return `#${scaleToMidRange(r).toString(16).padStart(2, '0')}${scaleToMidRange(
    g
  ).toString(16).padStart(2, '0')}${scaleToMidRange(b).toString(16).padStart(
    2,
    '0'
  )}`;
}
