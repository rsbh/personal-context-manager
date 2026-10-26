import { Callout, Flex, Text } from "@raystack/apsara";
import classes from "./ContextView.module.css";

interface NotFoundProps {
  id: string;
}

// NotFound is the callout for a context id that does not exist.
export function NotFound({ id }: NotFoundProps) {
  return (
    <Callout type="alert" variant="outline">
      <Flex direction="column" gap={2}>
        <Text weight="medium">Context not found</Text>
        <Text size="small" className={classes.path}>
          {id}
        </Text>
      </Flex>
    </Callout>
  );
}
