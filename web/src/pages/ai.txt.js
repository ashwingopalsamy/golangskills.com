const content = `# AI use and source declaration

Site: https://golangskills.com
Source: https://github.com/ashwingopalsamy/golangskills.com
License: Apache-2.0 (see the repository LICENSE for the governing terms)

The site serves public static catalog pages and documents. The canonical skill
instructions, metadata, evaluation cases, and evidence are maintained in the
repository. This file does not grant permissions beyond the project license.
`;

export function GET() {
  return new Response(content, { headers: { "Content-Type": "text/plain; charset=utf-8" } });
}
