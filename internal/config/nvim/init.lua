-- VibeTUI — LazyVim-based VS Code-like config
-- nvim is launched with NVIM_APPNAME=vibetui so this config lives at
-- ~/.config/vibetui/ and plugins install to ~/.local/share/vibetui/

local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    "git", "clone", "--filter=blob:none",
    "https://github.com/folke/lazy.nvim.git",
    "--branch=stable", lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

vim.g.mapleader        = " "
vim.g.maplocalleader   = "\\"
vim.g.lazyvim_picker   = "telescope"
vim.g.snacks_animate   = false
vim.g.autoformat       = false

local function detect_parsers()
  local cwd = vim.loop.cwd()
  local map = {
    ["*.go"] = "go",
    ["*.lua"] = "lua",
    ["*.py"] = "python",
    ["*.rs"] = "rust",
    ["*.js"] = "javascript",
    ["*.jsx"] = "javascript",
    ["*.ts"] = "typescript",
    ["*.tsx"] = "tsx",
    ["*.json"] = "json",
    ["*.yaml"] = "yaml",
    ["*.yml"] = "yaml",
    ["*.toml"] = "toml",
    ["*.md"] = "markdown",
    ["*.sh"] = "bash",
    ["Dockerfile"] = "dockerfile",
  }
  local seen = {
    bash = true,
    vim = true,
    lua = true,
    markdown = true,
    markdown_inline = true,
    query = true,
  }
  for pattern, parser in pairs(map) do
    local matches = vim.fn.globpath(cwd, "**/" .. pattern, false, true)
    if #matches > 0 then
      seen[parser] = true
    end
  end
  local out = {}
  for parser in pairs(seen) do
    table.insert(out, parser)
  end
  table.sort(out)
  return out
end

require("lazy").setup({
  { "LazyVim/LazyVim", import = "lazyvim.plugins" },

  {
    "folke/snacks.nvim",
    opts = {
      dashboard = { enabled = false },
    },
  },

  {
    "nvim-treesitter/nvim-treesitter",
    build = ":TSUpdate",
    opts = {
      ensure_installed = detect_parsers(),
      auto_install = true,
      highlight = { enable = true },
      indent = { enable = true },
    },
  },

  {
    "nvim-neo-tree/neo-tree.nvim",
    opts = {
      window = {
        position = "left",
        width = 30,
        mappings = {
          ["<cr>"]            = "open",
          ["<2-LeftMouse>"]   = "open",
          ["l"]               = "open",
          ["h"]               = "close_node",
          ["<space>"]         = "none",
        },
      },
      filesystem = {
        follow_current_file = { enabled = true },
        use_libuv_file_watcher = true,
        filtered_items = {
          visible = false,
          hide_dotfiles = false,
          hide_gitignored = false,
        },
      },
    },
    init = function()
      vim.api.nvim_create_autocmd("VimEnter", {
        callback = function()
          vim.defer_fn(function()
            if vim.fn.argc() == 0 then
              vim.cmd("enew")
            end
            require("neo-tree.command").execute({ action = "show", position = "left" })
          end, 10)
        end,
        once = true,
      })
    end,
  },

  {
    "akinsho/bufferline.nvim",
    opts = {
      options = {
        always_show_bufferline   = true,
        left_mouse_command       = "buffer %d",
        right_mouse_command      = function(n) require("snacks").bufdelete(n) end,
        middle_mouse_command     = function(n) require("snacks").bufdelete(n) end,
        close_command            = function(n) require("snacks").bufdelete(n) end,
        separator_style          = "slant",
        show_buffer_close_icons  = true,
        show_close_icon          = true,
        color_icons              = true,
        diagnostics              = "nvim_lsp",
        offsets = {
          {
            filetype   = "neo-tree",
            text       = "  Explorer",
            text_align = "left",
            separator  = true,
          },
        },
      },
    },
  },

  {
    "nvim-telescope/telescope.nvim",
    keys = {
      { "<C-p>",   "<cmd>Telescope find_files<cr>", desc = "Find Files" },
      { "<C-S-f>", "<cmd>Telescope live_grep<cr>",  desc = "Live Grep"  },
    },
    opts = {
      defaults = {
        layout_strategy = "horizontal",
        layout_config   = { horizontal = { preview_width = 0.55, width = 0.87, height = 0.80 } },
        sorting_strategy = "ascending",
        prompt_prefix   = "  ",
        selection_caret = " ",
        get_selection_window = function()
          for _, win in ipairs(vim.api.nvim_list_wins()) do
            if vim.bo[vim.api.nvim_win_get_buf(win)].buftype == "" then return win end
          end
          return 0
        end,
      },
    },
  },

  { "folke/which-key.nvim", opts = { delay = 500 } },

}, {
  defaults    = { lazy = true },
  install     = { colorscheme = { "tokyonight", "habamax" } },
  performance = {
    rtp = {
      disabled_plugins = {
        "gzip", "matchit", "matchparen", "netrwPlugin",
        "tarPlugin", "tohtml", "tutor", "zipPlugin",
      },
    },
  },
  ui = { border = "rounded" },
})

local opt = vim.opt
opt.mouse        = "a"
opt.mousemodel   = "popup_setpos"
opt.showtabline  = 2
opt.relativenumber = false
opt.number       = true
opt.guicursor    = "n-v-c:block,i-ci-ve:ver25,r-cr:hor20,o:hor50"
opt.autowrite    = true
opt.autowriteall = true
opt.updatetime   = 200

local autosave_group = vim.api.nvim_create_augroup("vibetui_autosave", { clear = true })
local modeless_group = vim.api.nvim_create_augroup("vibetui_modeless", { clear = true })

local function should_edit_buffer(bufnr)
  local bt = vim.bo[bufnr].buftype
  local ft = vim.bo[bufnr].filetype
  if bt ~= "" then
    return false
  end
  if ft == "neo-tree" or ft == "lazy" or ft == "TelescopePrompt" or ft == "snacks_picker_input" then
    return false
  end
  return true
end

local function autosave()
  local bufnr = vim.api.nvim_get_current_buf()
  if not should_edit_buffer(bufnr) then
    return
  end
  if vim.bo[bufnr].modifiable and not vim.bo[bufnr].readonly and vim.bo[bufnr].modified then
    pcall(vim.cmd, "silent! update")
  end
end

local function sync_insert_mode()
  local bufnr = vim.api.nvim_get_current_buf()
  local editable = should_edit_buffer(bufnr)
  vim.schedule(function()
    if vim.api.nvim_get_current_buf() ~= bufnr then
      return
    end
    local mode = vim.fn.mode()
    if editable then
      if mode ~= "i" then
        vim.cmd("startinsert")
      end
      return
    end
    if mode == "i" or mode == "ic" or mode == "ix" then
      vim.cmd("stopinsert")
    end
  end)
end

vim.api.nvim_create_autocmd({ "InsertLeave", "TextChanged", "TextChangedI", "FocusLost", "BufLeave" }, {
  group = autosave_group,
  callback = autosave,
})

vim.api.nvim_create_autocmd({ "BufEnter", "WinEnter" }, {
  group = modeless_group,
  callback = sync_insert_mode,
})

local map = vim.keymap.set
map("n",               "<C-s>",  "<cmd>w<cr>",                                  { desc = "Save" })
map("i",               "<C-s>",  "<C-o>:w<cr>",                                  { desc = "Save" })
map("v",               "<C-s>",  "<esc><cmd>w<cr>gv",                            { desc = "Save" })
map("n",               "<C-b>",  "<cmd>Neotree toggle<cr>",                    { desc = "Toggle Explorer" })
map("n",               "<C-`>",  function() require("snacks").terminal() end,  { desc = "Toggle Terminal" })
map("n",               "<Tab>",  "<cmd>BufferLineCycleNext<cr>",               { desc = "Next Buffer" })
map("n",               "<S-Tab>","<cmd>BufferLineCyclePrev<cr>",               { desc = "Prev Buffer" })
