/** The build identity this client reports when it joins a room. */
export const VERSION = '0.0.0-dev';

/**
 * Renders the version together with the role the client runs as. The control plane
 * keys on this string, so a blank role is reported rather than silently dropped.
 */
export function describe(role: string): string {
  const trimmed = role.trim();
  return `${trimmed === '' ? 'unknown' : trimmed}/${VERSION}`;
}
