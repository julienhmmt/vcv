<script lang="ts">
	import { LinkPreview as HoverCardPrimitive } from "bits-ui";

	let { ...restProps }: HoverCardPrimitive.PortalProps = $props();
</script>

<HoverCardPrimitive.Portal to="body" {...restProps} />
