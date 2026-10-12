export default {
  meta: {
    type: 'problem',
    schema: [],
    messages: {
      arbitrary:
        'Arbitrary Tailwind class syntax is banned. Define a named theme token, utility, or variant in style.css instead.',
    },
  },
  create(context) {
    const sourceCode = context.sourceCode;
    const classValues = [];
    const visited = new Set();

    function inspect(node) {
      if (!node || visited.has(node)) return;
      visited.add(node);

      const value =
        node.type === 'Literal' && typeof node.value === 'string'
          ? node.value
          : node.type === 'TemplateElement'
            ? (node.value.cooked ?? node.value.raw)
            : undefined;

      // Check fragments too, so interpolation cannot hide arbitrary syntax.
      if (value !== undefined && /\[|\]|-\(/.test(value)) {
        context.report({ node, messageId: 'arbitrary' });
      }

      // Follow local bindings such as Button's shared styles constant.
      if (node.type === 'Identifier') {
        let scope = sourceCode.getScope(node);
        while (scope) {
          const variable = scope.set.get(node.name);
          if (variable) {
            for (const definition of variable.defs) {
              if (definition.node.type === 'VariableDeclarator') {
                inspect(definition.node.init);
              }
            }
            break;
          }
          scope = scope.upper;
        }
        return;
      }

      for (const key of sourceCode.visitorKeys[node.type] ?? []) {
        const child = node[key];
        if (Array.isArray(child)) child.forEach(inspect);
        else inspect(child);
      }
    }

    return {
      JSXAttribute(node) {
        if (
          node.name.type === 'JSXIdentifier' &&
          node.name.name === 'className'
        ) {
          classValues.push(node.value);
        }
      },
      'Program:exit'() {
        classValues.forEach(inspect);
      },
    };
  },
};
