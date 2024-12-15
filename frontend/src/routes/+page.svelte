<script lang="ts">
    import {cn} from "$lib/utils";

    const puzzle: ('S' | 'U' | 'M' | 'X')[][] = [
        ['S', 'U', 'M'],
        ['U', 'S', 'U'],
        ['S', 'X', 'S']
    ];

    const sqSize: number = 20
</script>

<div class="flex h-screen w-screen justify-center items-center">
    <div class="grid grid-cols-2 grid-rows-2 place-items-center">
        <div></div>
        <div></div>
        <div></div>
        <div class="grid grid-cols-{puzzle.length} grid-rows-{puzzle.length} border-white border">
            {#each puzzle as row, i}
                {#each row as cell, j}
                    <div class={cn("size-10 bg-purple-400  place-content-center place-items-center aspect-square", "border-purple-200",
                                i !== 0 && "border-t-[0.5px]", j !== 0 && "border-l-[0.5px]",
                                i !== puzzle.length-1 && "border-b-[0.5px]", j !== puzzle.length-1 && "border-r-[0.5px]")}>
                        {#if cell === 'S'}
                            <div class="size-8 rounded bg-white aspect-square"></div>
                        {:else if cell === 'U'}
                            <div></div>
                        {:else if cell === 'M'}
                            <div class="size-8 border-2 rounded border-white aspect-square"></div>
                        {:else if cell === 'X'}
                            <div class="relative size-10 aspect-square">
                                <div class="absolute size-10 aspect-square">
                                    <div class="flex justify-center items-center w-full h-full">
                                        <div class="rotate-45 w-0.5 h-8 bg-black rounded-full"></div>
                                    </div>
                                </div>
                                <div class="absolute size-10 aspect-square">
                                    <div class="flex justify-center items-center w-full h-full">
                                        <div class="-rotate-45 w-0.5 h-8 bg-black rounded-full"></div>
                                    </div>
                                </div>
                            </div>
                        {/if}
                    </div>
                {/each}
            {/each}
        </div>
    </div>

</div>

<!-- grid-rows-3 grid-cols-3 -->
