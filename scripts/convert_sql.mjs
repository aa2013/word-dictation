/**
 * word.sql (MySQL dump) → SQLite 方言种子数据转换脚本
 *
 * 用法（工作区根目录执行）：
 *   node scripts/convert_sql.mjs [word.sql] [输出文件]
 *
 * 处理内容：
 *  1. 仅保留 INSERT 语句（表结构由迁移 0001 负责）
 *  2. MySQL 字符串转义 → SQLite 转义：
 *     \' → ''   \\ → \   \n/\r/\t → 真实控制字符
 *  3. 每条 INSERT 保持一条完整语句，供 Go 迁移器按引号感知方式逐条执行
 */
import fs from 'node:fs';
import path from 'node:path';

const src = process.argv[2] ?? 'word.sql';
const dst = process.argv[3] ?? 'server/migration/migrations/0002_seed_data.sql';

const raw = fs.readFileSync(src, 'utf8');
const lines = raw.split(/\r?\n/);

/** 把一行 MySQL INSERT 的字符串字面量转成 SQLite 字面量（逐字符状态机） */
function convertLine(line) {
  let out = '';
  let inStr = false;
  const unknown = [];
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (!inStr) {
      if (ch === "'") {
        inStr = true;
        out += "'";
      } else {
        out += ch;
      }
      continue;
    }
    // 字符串内部
    if (ch === '\\') {
      const n = line[++i];
      switch (n) {
        case "'": // MySQL 转义的单引号 → SQLite 用两个单引号
          out += "''";
          break;
        case '\\':
          out += '\\';
          break;
        case 'n':
          out += '\n';
          break;
        case 'r':
          out += '\r';
          break;
        case 't':
          out += '\t';
          break;
        case '"':
          out += '"';
          break;
        case '0':
          unknown.push('\\0');
          break;
        default:
          unknown.push('\\' + n);
          out += '\\' + n;
      }
      continue;
    }
    if (ch === "'") {
      inStr = false;
      out += "'";
      continue;
    }
    out += ch;
  }
  if (inStr) throw new Error('引号状态异常，转换中断: ' + line.slice(0, 80));
  return { sql: out, unknown };
}

const counts = {};
const outLines = [];
let unknownTotal = [];
let converted = 0;

for (const line of lines) {
  const t = line.trim();
  if (!t.startsWith('INSERT INTO')) continue;
  const table = t.match(/^INSERT INTO `([a-z_]+)`/)?.[1];
  if (!table) throw new Error('无法解析表名: ' + t.slice(0, 60));
  const { sql, unknown } = convertLine(t);
  if (unknown.length) unknownTotal.push(...unknown);
  counts[table] = (counts[table] ?? 0) + 1;
  outLines.push(sql);
  converted++;
}

const header = `-- ------------------------------------------------------------
-- 0002_seed_data.sql
-- 由 scripts/convert_sql.mjs 从 word.sql (MySQL dump) 自动转换生成
-- 仅数据（表结构见 0001_init.sql）
-- ------------------------------------------------------------`;

fs.mkdirSync(path.dirname(dst), { recursive: true });
fs.writeFileSync(dst, header + '\n' + outLines.join('\n') + '\n', 'utf8');

console.log('转换完成 →', dst);
console.log('语句总数:', converted);
console.log('各表行数:', counts);
console.log('未知转义:', unknownTotal.length ? unknownTotal : '无');
console.log('文件大小:', (fs.statSync(dst).size / 1024).toFixed(1), 'KB');
