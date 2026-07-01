import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
import matplotlib.patheffects as pe
from matplotlib.patches import FancyBboxPatch, FancyArrowPatch
import numpy as np

# ─────────────────────────────────────────────
#  Global variables
# ─────────────────────────────────────────────
SLIDE_W, SLIDE_H = 16, 9          # inches  (16:9)
DPI               = 200
BG_COLOR          = "#0f1923"      # dark navy
ACCENT1           = "#00c9ff"      # cyan
ACCENT2           = "#f97316"      # orange
ACCENT3           = "#a3e635"      # lime
ACCENT4           = "#c084fc"      # purple
TEXT_LIGHT        = "#e2e8f0"
TEXT_DIM          = "#94a3b8"
GENE_COLORS       = {0: "#1e3a5f", 1: "#00c9ff"}  # off / on
FONT_TITLE        = dict(fontsize=28, fontweight="bold", color=TEXT_LIGHT,
                         fontfamily="monospace")
FONT_SUB          = dict(fontsize=18, color=TEXT_DIM, fontfamily="monospace")
FONT_BODY         = dict(fontsize=16, color=TEXT_LIGHT, fontfamily="monospace")
FONT_LABEL        = dict(fontsize=14, color=TEXT_LIGHT, fontfamily="monospace")
FONT_SMALL        = dict(fontsize=12, color=TEXT_DIM,   fontfamily="monospace")

EXAMPLE_GENOME    = [1, 0, 1, 1, 0, 0, 1]
PARENT_A          = [0, 1, 0, 0, 1, 1, 0]
PARENT_B          = [1, 0, 1, 0, 0, 1, 1]
CROSSOVER_PT      = 4             # genes 0-3 from A, 4-6 from B


def new_slide(title=None, subtitle=None):
    """Create a blank dark slide and return (fig, ax)."""
    fig, ax = plt.subplots(figsize=(SLIDE_W, SLIDE_H), dpi=DPI)
    fig.patch.set_facecolor(BG_COLOR)
    ax.set_facecolor(BG_COLOR)
    ax.set_xlim(0, 1)
    ax.set_ylim(0, 1)
    ax.axis("off")

    # thin top accent bar
    ax.axhline(0.96, color=ACCENT1, linewidth=2.5, xmin=0.04, xmax=0.96)

    if title:
        ax.text(0.5, 0.88, title, ha="center", va="center", **FONT_TITLE)
    if subtitle:
        ax.text(0.5, 0.82, subtitle, ha="center", va="center", **FONT_SUB)
    return fig, ax


def save_slide(fig, name, has_inset=False):
    """Save slide; skip tight_layout when inset axes are present."""
    #if not has_inset:
    fig.tight_layout()
    fig.savefig(name, dpi=DPI, bbox_inches="tight",
                facecolor=fig.get_facecolor())
    plt.close(fig)
    print(f"  saved -> {name}")


# ─────────────────────────────────────────────
#  Helper: draw a genome strip
# ─────────────────────────────────────────────
def draw_genome(ax, genome, cx, cy, cell_w=0.055, cell_h=0.09,
                label=None, highlight=None, alpha=1.0, fontsize=16):
    """
    Draw a genome as a row of coloured cells centred at (cx, cy).
    highlight: list of indices to outline in ACCENT2.
    """
    n = len(genome)
    total_w = n * cell_w
    x0 = cx - total_w / 2

    for i, gene in enumerate(genome):
        x = x0 + i * cell_w
        y = cy - cell_h / 2
        facecolor = GENE_COLORS[gene]
        ec = ACCENT2 if (highlight and i in highlight) else "#ffffff33"
        lw = 3 if (highlight and i in highlight) else 0.8
        rect = FancyBboxPatch((x + 0.003, y + 0.003),
                              cell_w - 0.006, cell_h - 0.006,
                              boxstyle="round,pad=0.004",
                              facecolor=facecolor, edgecolor=ec,
                              linewidth=lw, alpha=alpha,
                              transform=ax.transAxes, clip_on=False)
        ax.add_patch(rect)
        ax.text(x + cell_w / 2, cy, str(gene),
                ha="center", va="center",
                fontsize=fontsize, fontweight="bold",
                color=TEXT_LIGHT if gene else TEXT_DIM,
                fontfamily="monospace",
                transform=ax.transAxes, alpha=alpha)

    if label:
        ax.text(x0 - 0.015, cy, label, ha="right", va="center",
                **FONT_LABEL, transform=ax.transAxes, alpha=alpha)
    return x0, total_w


# ═══════════════════════════════════════════════════════════════
#  SLIDE 1 — Title slide
# ═══════════════════════════════════════════════════════════════
def slide_title():
    fig, ax = new_slide()

    ax.text(0.5, 0.62, "Genetic Algorithms",
            ha="center", va="center",
            fontsize=46, fontweight="bold",
            color=ACCENT1, fontfamily="monospace")
    ax.text(0.5, 0.50, "Optimization inspired by natural evolution",
            ha="center", va="center", **FONT_SUB)

    # decorative genome strip
    #genome = [1, 0, 1, 1, 0, 0, 1, 0, 1, 1, 0, 1]
    #draw_genome(ax, genome, cx=0.5, cy=0.36, cell_w=0.055, cell_h=0.08,
    #            fontsize=15)

    ax.text(0.5, 0.12, "Selection  ->  Crossover  ->  Mutation  ->  Next Generation",
            ha="center", va="center", **FONT_BODY)

    save_slide(fig, "slide_01_title.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 2 — Solution representation
# ═══════════════════════════════════════════════════════════════
def slide_representation():
    fig, ax = new_slide(
        title="Solution Representation",
        subtitle="Each candidate solution is encoded as a chromosome"
    )

    # stick figure
    head = plt.Circle((0.18, 0.56), 0.045, color=ACCENT1,
                       transform=ax.transAxes, zorder=3)
    ax.add_patch(head)
    ax.plot([0.18, 0.18], [0.51, 0.36], color=ACCENT1, lw=3,
            transform=ax.transAxes)
    ax.plot([0.12, 0.24], [0.46, 0.46], color=ACCENT1, lw=3,
            transform=ax.transAxes)
    ax.plot([0.18, 0.12], [0.36, 0.25], color=ACCENT1, lw=3,
            transform=ax.transAxes)
    ax.plot([0.18, 0.24], [0.36, 0.25], color=ACCENT1, lw=3,
            transform=ax.transAxes)
    ax.text(0.18, 0.20, "Individual", ha="center", **FONT_LABEL,
            transform=ax.transAxes)

    ax.annotate("", xy=(0.35, 0.42), xytext=(0.28, 0.42),
                arrowprops=dict(arrowstyle="-|>", color=TEXT_DIM, lw=1.5),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(0.315, 0.45, "encoded as", ha="center", **FONT_SMALL,
            transform=ax.transAxes)

    draw_genome(ax, EXAMPLE_GENOME, cx=0.62, cy=0.42,
                cell_w=0.065, cell_h=0.10, fontsize=17)

    ax.annotate("", xy=(0.75, 0.48), xytext=(0.75, 0.62),
                arrowprops=dict(arrowstyle="-|>", color=ACCENT2, lw=1.5),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(0.75, 0.65, "one gene\n(bit = 0 or 1)",
            ha="center", va="bottom",
            fontsize=11, color=ACCENT2, fontfamily="monospace",
            transform=ax.transAxes)

    save_slide(fig, "slide_02_representation.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 3 — Fitness function
# ═══════════════════════════════════════════════════════════════
def slide_fitness():
    fig, ax = new_slide(
        title="Fitness Function",
        subtitle="f(x) = number of 1s in the genome   ->   higher = better"
    )

    genomes = [
        [0, 0, 1, 0, 0, 0, 1],
        [1, 0, 1, 0, 0, 1, 0],
        [1, 1, 1, 0, 1, 0, 1],
        [1, 1, 1, 1, 1, 0, 1],
        [1, 1, 1, 1, 1, 1, 1],
    ]
    scores = [sum(g) for g in genomes]
    colors = [ACCENT1 if s == max(scores) else "#1e4e6b" for s in scores]

    # inset axes for bar chart
    inset = fig.add_axes([0.5, 0.22, 0.35, 0.45])
    inset.set_facecolor(BG_COLOR)
    inset.barh(range(len(scores)), scores,
               color=colors, edgecolor="#130f0f21", height=0.55)
    inset.set_yticks(range(len(scores)))
    inset.set_yticklabels(
        [str(g).replace(" ", "") for g in genomes],
        fontsize=9, fontfamily="monospace", color=TEXT_LIGHT
    )
    inset.set_xlabel("Fitness  f(x)", color=TEXT_DIM,
                     fontfamily="monospace", fontsize=11)
    inset.set_xlim(0, 8)
    inset.spines[["top", "right"]].set_visible(False)
    for sp in ["bottom", "left"]:
        inset.spines[sp].set_color(TEXT_DIM)
    inset.tick_params(colors=TEXT_DIM, labelsize=10)
    inset.xaxis.label.set_color(TEXT_DIM)

    for i, v in enumerate(scores):
        inset.text(v + 0.1, i, str(v), va="center",
                   fontsize=11,
                   color=ACCENT1 if v == max(scores) else TEXT_DIM,
                   fontfamily="monospace")

    inset.annotate("best individual", xy=(7, 4), xytext=(5.5, 3.3),
                   arrowprops=dict(arrowstyle="-|>", color=ACCENT2, lw=1.3),
                   fontsize=10, color=ACCENT2, fontfamily="monospace")

    # pipeline boxes on the left
    for step_i, label in enumerate([
        "genome\n[1,0,1,1,0,0,1]",
        "fitness\nfunction  f(x)",
        "score\nf = 4",
    ]):
        y = 0.65 - step_i * 0.20
        box = FancyBboxPatch((0.06, y - 0.06), 0.28, 0.12,
                             boxstyle="round,pad=0.01",
                             facecolor="#112233", edgecolor=ACCENT1,
                             linewidth=1.5, transform=ax.transAxes)
        ax.add_patch(box)
        ax.text(0.20, y, label, ha="center", va="center",
                fontsize=11, color=TEXT_LIGHT, fontfamily="monospace",
                transform=ax.transAxes)
        if step_i < 2:
            ax.annotate("", xy=(0.20, y - 0.08), xytext=(0.20, y - 0.13),
                        arrowprops=dict(arrowstyle="<|-", color=ACCENT1, lw=1.5),
                        xycoords="axes fraction", textcoords="axes fraction")

    # has_inset=True suppresses tight_layout to avoid the UserWarning
    save_slide(fig, "slide_03_fitness.png", has_inset=True)


# ═══════════════════════════════════════════════════════════════
#  SLIDE 4 — Population & Generations
# ═══════════════════════════════════════════════════════════════
def slide_population():
    fig, ax = new_slide(
        title="Population & Generations",
        subtitle="A population evolves over many generations"
    )

    rng = np.random.default_rng(42)

    def random_genome():
        return list(rng.integers(0, 2, 7))

    n_gen = 3
    n_ind = 5
    gen_x = [0.15, 0.50, 0.85]
    ind_ys = np.linspace(0.6, 0.16, n_ind)

    for gi, gx in enumerate(gen_x):
        box = FancyBboxPatch((gx - 0.13, 0.08), 0.26, 0.62,
                             boxstyle="round,pad=0.01",
                             facecolor="#0a1628",
                             edgecolor=ACCENT1 if gi == 0 else TEXT_DIM,
                             linewidth=2 if gi == 0 else 0.8,
                             transform=ax.transAxes)
        ax.add_patch(box)
        ax.text(gx, 0.74, f"Generation {gi + 1}",
                ha="center", va="center",
                fontsize=12, fontweight="bold",
                color=ACCENT1 if gi == 0 else TEXT_DIM,
                fontfamily="monospace", transform=ax.transAxes)

        for iy, gy in enumerate(ind_ys):
            g = random_genome()
            if gi == 1:
                g = [1 if (v == 1 or rng.random() < 0.25) else 0 for v in g]
            elif gi == 2:
                g = [1 if (v == 1 or rng.random() < 0.45) else 0 for v in g]
            score = sum(g)
            draw_genome(ax, g, cx=gx, cy=gy,
                        cell_w=0.032, cell_h=0.06, fontsize=10)
            ax.text(gx + 0.115, gy, f"f={score}",
                    ha="left", va="center",
                    fontsize=9,
                    color=ACCENT3 if score >= 5 else TEXT_DIM,
                    fontfamily="monospace", transform=ax.transAxes)

    for i in range(n_gen - 1):
        ax.annotate("", xy=(gen_x[i + 1] - 0.14, 0.48),
                    xytext=(gen_x[i] + 0.14, 0.48),
                    arrowprops=dict(arrowstyle="-|>", color=ACCENT2, lw=2.5),
                    xycoords="axes fraction", textcoords="axes fraction")
        ax.text((gen_x[i] + gen_x[i + 1]) / 2, 0.54,
                "evolve", ha="center",
                fontsize=10, color=ACCENT2, fontfamily="monospace",
                transform=ax.transAxes)

    save_slide(fig, "slide_04_population.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 5 — Tournament Selection
# ═══════════════════════════════════════════════════════════════
def slide_selection():
    fig, ax = new_slide(
        title="Selection - Tournament",
        subtitle="Pick  k  random individuals  ->  the fittest one becomes a parent"
    )

    rng = np.random.default_rng(7)
    pool_genomes = [list(rng.integers(0, 2, 7)) for _ in range(6)]
    pool_scores  = [sum(g) for g in pool_genomes]
    tournament   = [1, 3, 4]
    winner       = max(tournament, key=lambda i: pool_scores[i])

    xs = [0.18, 0.50, 0.82, 0.18, 0.50, 0.82]
    ys = [0.65, 0.65, 0.65, 0.38, 0.38, 0.38]

    for idx, (gx, gy) in enumerate(zip(xs, ys)):
        in_tournament = idx in tournament
        is_winner     = idx == winner

        alpha = 1.0 if in_tournament else 0.30
        ec = ACCENT2 if is_winner else (ACCENT1 if in_tournament else TEXT_DIM)
        lw = 3 if is_winner else (1.5 if in_tournament else 0.5)

        box = FancyBboxPatch((gx - 0.12, gy - 0.08), 0.24, 0.16,
                             boxstyle="round,pad=0.01",
                             facecolor="#0a1628",
                             edgecolor=ec, linewidth=lw, alpha=alpha,
                             transform=ax.transAxes)
        ax.add_patch(box)
        draw_genome(ax, pool_genomes[idx], cx=gx, cy=gy + 0.01,
                    cell_w=0.028, cell_h=0.055, fontsize=9, alpha=alpha)
        ax.text(gx, gy - 0.055,
                f"f = {pool_scores[idx]}",
                ha="center", fontsize=11,
                color=ACCENT2 if is_winner else (ACCENT1 if in_tournament else TEXT_DIM),
                fontweight="bold" if is_winner else "normal",
                fontfamily="monospace", transform=ax.transAxes, alpha=alpha)

        if is_winner:
            # "BEST" text badge instead of trophy emoji
            ax.text(gx, gy + 0.10, "[ WINNER ]",
                    ha="center", fontsize=12, color=ACCENT2,
                    fontweight="bold",
                    fontfamily="monospace", transform=ax.transAxes)

    for bx, col, lbl in [(0.25, ACCENT2,  "winner"),
                          (0.43, ACCENT1,  "tournament pool"),
                          (0.66, TEXT_DIM, "not selected")]:
        ax.plot(bx, 0.10, "s", color=col, ms=10,
                transform=ax.transAxes)
        ax.text(bx + 0.025, 0.10, lbl, va="center",
                **FONT_SMALL, transform=ax.transAxes)

    save_slide(fig, "slide_05_selection.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 6 — Crossover
# ═══════════════════════════════════════════════════════════════
def slide_crossover():
    fig, ax = new_slide(
        title="Reproduction - Crossover",
        subtitle="Combine two parents to create offspring"
    )

    cell_w, cell_h = 0.065, 0.10
    n = len(PARENT_A)

    rows = [
        (PARENT_A, 0.65, "Parent A", list(range(CROSSOVER_PT))),
        (PARENT_B, 0.48, "Parent B", list(range(CROSSOVER_PT, n))),
    ]

    offspring = PARENT_A[:CROSSOVER_PT] + PARENT_B[CROSSOVER_PT:]

    for genome, cy, label, hl in rows:
        draw_genome(ax, genome, cx=0.50, cy=cy,
                    cell_w=cell_w, cell_h=cell_h,
                    label=label, highlight=hl, fontsize=16)

    total_w = n * cell_w
    x0 = 0.50 - total_w / 2
    cut_x = x0 + CROSSOVER_PT * cell_w
    ax.plot([cut_x, cut_x], [0.43, 0.70],
            color=ACCENT2, lw=2.5, linestyle="--",
            transform=ax.transAxes)
    ax.text(cut_x, 0.72, "cut point",
            ha="center", fontsize=11, color=ACCENT2,
            fontfamily="monospace", transform=ax.transAxes)

    ax.annotate("", xy=(0.50, 0.34), xytext=(0.50, 0.42),
                arrowprops=dict(arrowstyle="-|>", color=ACCENT3, lw=2.5),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(0.51, 0.38, "combine", ha="left",
            fontsize=11, color=ACCENT3, fontfamily="monospace",
            transform=ax.transAxes)

    draw_genome(ax, offspring, cx=0.50, cy=0.27,
                cell_w=cell_w, cell_h=cell_h,
                label="Offspring", fontsize=16)

    save_slide(fig, "slide_06_crossover.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 7 — Mutation
# ═══════════════════════════════════════════════════════════════
def slide_mutation():
    fig, ax = new_slide(
        title="Reproduction - Mutation",
        subtitle="A random gene is flipped  ->  maintains diversity"
    )

    before  = [1, 0, 1, 1, 0, 0, 1]
    mut_idx = 4
    after   = before.copy()
    after[mut_idx] = 1 - after[mut_idx]

    cell_w, cell_h = 0.075, 0.11

    draw_genome(ax, before, cx=0.50, cy=0.65,
                cell_w=cell_w, cell_h=cell_h, label="Before", fontsize=18)
    draw_genome(ax, after,  cx=0.50, cy=0.38,
                cell_w=cell_w, cell_h=cell_h, label="After ",
                highlight=[mut_idx], fontsize=18)

    ax.annotate("", xy=(0.50, 0.46), xytext=(0.50, 0.57),
                arrowprops=dict(arrowstyle="-|>", color=ACCENT2, lw=3),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(0.44, 0.515, "mutate\ngene 4", ha="left", va="center",
            fontsize=12, color=ACCENT2, fontfamily="monospace",
            transform=ax.transAxes)

    total_w = len(after) * cell_w
    x0 = 0.50 - total_w / 2
    gene_cx = x0 + mut_idx * cell_w + cell_w / 2
    ax.annotate("",
                xy=(gene_cx, 0.38 + cell_h / 2 + 0.02),
                xytext=(gene_cx, 0.38 + cell_h / 2 + 0.08),
                arrowprops=dict(arrowstyle="-|>", color=ACCENT3, lw=2),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(gene_cx, 0.38 + cell_h / 2 + 0.1,
            "0  ->  1\n(flipped!)",
            ha="center", va="bottom",
            fontsize=12, color=ACCENT3, fontfamily="monospace",
            transform=ax.transAxes)

    save_slide(fig, "slide_07_mutation.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 8 — Elitism
# ═══════════════════════════════════════════════════════════════
def slide_elitism():
    fig, ax = new_slide(
        title="Elitism",
        subtitle="The best N individuals survive unchanged into the next generation"
    )

    rng = np.random.default_rng(3)
    pop = [list(rng.integers(0, 2, 7)) for _ in range(4)]
    scores = [sum(g) for g in pop]
    elite_n = 2
    order = sorted(range(len(pop)), key=lambda i: -scores[i])
    elite_set = set(order[:elite_n])

    ys_left  = np.linspace(0.65, 0.22, len(pop))
    ys_right = np.linspace(0.65, 0.22, len(pop))

    ax.text(0.20, 0.72, "Current Generation",
            ha="center", fontsize=13, fontweight="bold",
            color=TEXT_DIM, fontfamily="monospace",
            transform=ax.transAxes)
    ax.text(0.78, 0.72, "Next Generation",
            ha="center", fontsize=13, fontweight="bold",
            color=TEXT_DIM, fontfamily="monospace",
            transform=ax.transAxes)

    for rank, idx in enumerate(order):
        gy = ys_left[rank]
        is_elite = idx in elite_set
        draw_genome(ax, pop[idx], cx=0.22, cy=gy,
                    cell_w=0.042, cell_h=0.075, fontsize=12)
        if is_elite:
            # star marker instead of emoji
            ax.text(0.05, gy, "(*)",
                    ha="center", va="center",
                    fontsize=14, color=ACCENT2,
                    fontfamily="monospace",
                    transform=ax.transAxes)

    for rank_r in range(len(pop)):
        gy = ys_right[rank_r]
        if rank_r < elite_n:
            src_idx = order[rank_r]
            draw_genome(ax, pop[src_idx], cx=0.78, cy=gy,
                        cell_w=0.042, cell_h=0.075, fontsize=12)
            ax.annotate("", xy=(0.56, gy), xytext=(0.46, gy),
                        arrowprops=dict(arrowstyle="-|>",
                                       color=ACCENT2, lw=2.0),
                        xycoords="axes fraction",
                        textcoords="axes fraction")
        else:
            for bx_i in range(7):
                bx = 0.78 - (7 * 0.042) / 2 + bx_i * 0.042
                rect = FancyBboxPatch((bx + 0.002, gy - 0.033),
                                     0.036, 0.065,
                                     boxstyle="round,pad=0.003",
                                     facecolor="#0a1628",
                                     edgecolor="#ffffff22", linewidth=0.6,
                                     transform=ax.transAxes)
                ax.add_patch(rect)
            ax.text(0.78, gy, "offspring  (crossover / mutation)",
                    ha="center", va="center", fontsize=9,
                    color=TEXT_DIM, fontfamily="monospace",
                    transform=ax.transAxes)

    save_slide(fig, "slide_08_elitism.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 9 — Speciation & Hamming Distance
# ═══════════════════════════════════════════════════════════════
def slide_speciation():
    fig, ax = new_slide(
        title="Speciation & Fitness Sharing",
        subtitle="Group similar individuals into species  ->  encourage diversity"
    )

    gA = [0, 1, 0, 0, 1, 1, 0]
    gB = [0, 0, 1, 0, 1, 1, 1]
    gC = [1, 1, 1, 1, 0, 0, 0]

    cell_w, cell_h = 0.052, 0.085

    # ── Hamming distance (left half) ────────────────────────────
    ax.text(0.27, 0.75, "Hamming Distance",
            ha="center", fontsize=14, fontweight="bold",
            color=ACCENT1, fontfamily="monospace",
            transform=ax.transAxes)

    diff_AB = [i for i in range(len(gA)) if gA[i] != gB[i]]
    draw_genome(ax, gA, cx=0.27, cy=0.65, cell_w=cell_w, cell_h=cell_h,
                label="A", highlight=diff_AB, fontsize=14)
    draw_genome(ax, gB, cx=0.27, cy=0.51, cell_w=cell_w, cell_h=cell_h,
                label="B", highlight=diff_AB, fontsize=14)
    ax.text(0.27, 0.43,
            f"H(A,B) = {len(diff_AB)}  -> same species  (threshold = 3)",
            ha="center", fontsize=11, color=ACCENT3,
            fontfamily="monospace", transform=ax.transAxes)

    diff_AC = [i for i in range(len(gA)) if gA[i] != gC[i]]
    draw_genome(ax, gA, cx=0.27, cy=0.31, cell_w=cell_w, cell_h=cell_h,
                label="A", highlight=diff_AC, fontsize=14)
    draw_genome(ax, gC, cx=0.27, cy=0.17, cell_w=cell_w, cell_h=cell_h,
                label="C", highlight=diff_AC, fontsize=14)
    ax.text(0.27, 0.09,
            f"H(A,C) = {len(diff_AC)}  -> different species",
            ha="center", fontsize=11, color=ACCENT2,
            fontfamily="monospace", transform=ax.transAxes)

    # ── vertical divider (using ax.plot instead of axvline) ─────
    ax.plot([0.55, 0.55], [0.08, 0.75],
            color=TEXT_DIM, linewidth=0.8,
            transform=ax.transAxes)

    # ── Fitness sharing (right half) ────────────────────────────
    ax.text(0.775, 0.75, "Fitness Sharing",
            ha="center", fontsize=14, fontweight="bold",
            color=ACCENT4, fontfamily="monospace",
            transform=ax.transAxes)

    species = [
        {"label": "Species 1", "members": 4, "raw_f": 6, "cx": 0.65},
        {"label": "Species 2", "members": 1, "raw_f": 5, "cx": 0.90},
    ]
    for si, sp in enumerate(species):
        cy_base = 0.63
        col = ACCENT1 if si == 0 else ACCENT4
        ax.text(sp["cx"], cy_base + 0.07, sp["label"],
                ha="center", fontsize=12, fontweight="bold",
                color=col, fontfamily="monospace",
                transform=ax.transAxes)

        for mi in range(sp["members"]):
            dot_x = sp["cx"] #+ (mi - sp["members"] / 2) * 0.06
            ax.plot(dot_x, cy_base - mi * 0.07, "o",
                    markersize=12, color=col, alpha=0.7,
                    transform=ax.transAxes)

        raw    = sp["raw_f"]
        size   = sp["members"]
        shared = raw / size
        ax.text(sp["cx"], cy_base - size * 0.07 - 0.04,
                f"f  = {raw}",
                ha="center", fontsize=11, color=TEXT_LIGHT,
                fontfamily="monospace", transform=ax.transAxes)
        ax.text(sp["cx"], cy_base - size * 0.07 - 0.12,
                f"|s| = {size}",
                ha="center", fontsize=11, color=TEXT_LIGHT,
                fontfamily="monospace", transform=ax.transAxes)
        ax.text(sp["cx"], cy_base - size * 0.07 - 0.20,
                f"f' = {shared:.1f}",
                ha="center", fontsize=13, fontweight="bold",
                color=ACCENT3, fontfamily="monospace",
                transform=ax.transAxes)

    ax.text(0.775, 0.10,
            "f'(i) = f(i) / |s(i)|",
            ha="center", fontsize=13, color=ACCENT3,
            fontfamily="monospace", transform=ax.transAxes)

    save_slide(fig, "slide_09_speciation.png")


# ═══════════════════════════════════════════════════════════════
#  SLIDE 10 — Full GA Flow Diagram
# ═══════════════════════════════════════════════════════════════
def slide_flowchart():
    fig, ax = new_slide(
        title="Genetic Algorithm - Overview",
        subtitle=""
    )

    steps = [
        ("Initialise\nPopulation",  0.50, 0.78, ACCENT1),
        ("Evaluate\nFitness",       0.50, 0.65, ACCENT1),
        ("Tournament\nSelection",   0.50, 0.52, ACCENT1),
        ("Crossover\n& Mutation",   0.50, 0.39, ACCENT1),
        ("Elitism\n+ New Pop.",     0.50, 0.26, ACCENT1),
        ("Converged?",              0.50, 0.13, ACCENT2),
    ]

    box_w, box_h = 0.22, 0.07

    for i, (label, bx, by, col) in enumerate(steps):
        rect = FancyBboxPatch((bx - box_w / 2, by - box_h / 2),
                              box_w, box_h,
                              boxstyle="round,pad=0.01",
                              facecolor="#0d2137",
                              edgecolor=col, linewidth=2.5,
                              transform=ax.transAxes)
        ax.add_patch(rect)
        ax.text(bx, by, label, ha="center", va="center",
                fontsize=12, fontweight="bold",
                color=TEXT_LIGHT, fontfamily="monospace",
                transform=ax.transAxes)

        if i < len(steps) - 1:
            next_by = steps[i + 1][2]
            ax.annotate("", xy=(bx, next_by + box_h / 2 + 0.005),
                        xytext=(bx, by - box_h / 2 - 0.005),
                        arrowprops=dict(arrowstyle="-|>",
                                       color=TEXT_DIM, lw=1.8),
                        xycoords="axes fraction",
                        textcoords="axes fraction")

    # "No" loop back arrow along the left side
    ax.annotate("", xy=(0.20, 0.65), xytext=(0.38, 0.65),
                arrowprops=dict(arrowstyle="<|-", color=ACCENT3, lw=2.0),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.plot([0.2, 0.20, 0.38], [0.65, 0.13, 0.13],
            color=ACCENT3, lw=2, transform=ax.transAxes)
    ax.text(0.14, 0.39, "No\n(loop)", ha="center",
            fontsize=11, color=ACCENT3, fontfamily="monospace",
            transform=ax.transAxes)

    # "Yes" exit arrow to the right
    ax.annotate("", xy=(0.7, 0.13), xytext=(0.62, 0.13),
                arrowprops=dict(arrowstyle="-|>", color=ACCENT2, lw=2),
                xycoords="axes fraction", textcoords="axes fraction")
    ax.text(0.72, 0.13, "Yes ->\nReturn best\nindividual",
            ha="left", va="center",
            fontsize=11, color=ACCENT2, fontfamily="monospace",
            transform=ax.transAxes)

    # side annotations
    side_notes = [
        (0.68, 0.78, "<- random bit-strings"),
        (0.68, 0.65, "<- f(x) = sum of genes"),
        (0.68, 0.52, "<- k random, best wins"),
        (0.68, 0.39, "<- one-point cut / bit-flip"),
        (0.68, 0.26, "<- keep top N unchanged"),
    ]
    for sx, sy, note in side_notes:
        ax.text(sx, sy, note, va="center",
                fontsize=9, color=TEXT_DIM, fontfamily="monospace",
                transform=ax.transAxes)

    save_slide(fig, "slide_10_flowchart.png")


# ─────────────────────────────────────────────
#  Run all slides
# ─────────────────────────────────────────────
if __name__ == "__main__":
    print("Generating slides...")
    slide_title()
    slide_representation()
    slide_fitness()
    slide_population()
    slide_selection()
    slide_crossover()
    slide_mutation()
    slide_elitism()
    slide_speciation()
    slide_flowchart()
    print("Done -- 10 slides saved.")

    from PIL import Image
    import glob
    import re
    import os

    # Get all matching files and sort by number
    image_files = glob.glob("slide_*.png")
    image_files.sort(key=lambda f: int(re.search(r'slide_(\d+)_', f).group(1)))

    # Convert and save as PDF
    images = [Image.open(f).convert("RGB") for f in image_files]
    images[0].save(
        "presentation.pdf",
        save_all=True,
        append_images=images[1:]
    )

    print(f"Created PDF with {len(images)} slides")