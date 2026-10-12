import { createCn } from 'cn/config';

// Keep custom theme tokens in sync with style.css for conflict resolution.
export const cn = createCn({
  extend: {
    theme: {
      text: ['label', 'meta', 'body', 'brand'],
      'font-weight': ['label', 'brand'],
      radius: ['control', 'panel'],
      shadow: ['layout'],
      spacing: [
        'control-y',
        'control-mobile-x',
        'toolbar',
        'repository',
        'sidebar',
      ],
      tracking: ['label', 'brand'],
    },
  },
});
